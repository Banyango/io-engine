package ecs

import (
	"encoding/json"
	"fmt"
	"github.com/goburrow/dynamic"
	"reflect"
	"strconv"
	"sync"
)

type Component interface {
	Id() int
	CreateComponent()
	DestroyComponent()
	Clone() Component
}

type CompareComponent interface {
	AreEquals(component Component) bool
}

type ComponentType int

const (
	PositionComponentType ComponentType = iota
	CollisionComponentType
	StateComponentType
	CircleComponentType
	ArcadeMovementComponentType
	NetworkInstanceComponentType
)

type Storage struct {
	Components map[int64]*Component
}

func NewStorage() Storage {
	store := Storage{}

	store.Components = make(map[int64]*Component)

	return store
}

/**
Entities
*/

type Entity struct {
	Id         int64             `json:"id"`
	Components map[int]Component `json:"components"`
}

func (entity Entity) Clone() *Entity {

	result := Entity{}
	result.Id = entity.Id

	result.Components = map[int]Component{}

	for i, val := range entity.Components {
		result.Components[i] = val.Clone()
	}

	return &result
}

func (entity *Entity) CompareTo(other *Entity) (same bool){

	for i, comp := range entity.Components {
		if val, ok := comp.(CompareComponent); ok {
			if !val.AreEquals(other.Components[i]) {
				return false
			}
		}
	}

	return true
}

func NewEntity() Entity {
	e := Entity{}
	e.Components = make(map[int]Component)
	return e
}

/**
Systems are behaviors

Systems store a list of components
*/
type System interface {
	Init(w *World)
	AddToStorage(entity *Entity)
	RequiredComponentTypes() []ComponentType
	UpdateSystem(delta float64, world *World)
	RemoveFromStorage(entity *Entity)
}

/**
World

This is the top level container for all logic.

Systems  - Store behaviours and have storages.
Entities - Have components that get added to storages.
Storages - Have components that are a slice of all components needed.
Globals  - Are static components like input that dont really belong
		   to one entity in particular.
*/

const (
	FIXED_DELTA = 0.016
	MAX_CACHE_SIZE = 32
)

type World struct {
	IdIndex int64

	Systems       []*System
	RenderSystems []*System

	Entities map[int64]*Entity

	Cache []map[int64]*Entity
	CacheInput []InputController
	ValidatedBuffer int32
	IsResimulating bool

	ToSpawn []Entity
	ToDestroy []int64

	Input InputController

	TimeElapsed      int64
	LastFrameTime    int64
	CurrentFrameTime int64
	CurrentTick      int64

	Interval   int64
	PrefabData *PrefabData
	mux        sync.Mutex
}

func (w *World) Update(delta float64) {
	w.CurrentTick++

	for _, v := range w.Systems {
		(*v).UpdateSystem(delta, w)
	}

	w.CacheState()
}

func (self *World) Render() {
	for _, v := range self.RenderSystems {
		(*v).UpdateSystem(-1, self)
	}
}

func NewWorld() *World {
	world := new(World)

	world.Entities = map[int64]*Entity{}
	world.Input = InputController{map[PlayerId]*Input{0: NewInput()}}

	world.Interval = 16

	return world
}

func (w *World) AddSystem(system System) {
	owned := &system
	w.Systems = append(w.Systems, owned)
	fmt.Println("Adding System: ", reflect.TypeOf(system))
	(*owned).Init(w)
	fmt.Println("initialized: ", reflect.TypeOf(system))
}

func (w *World) AddRenderer(system System) {
	owned := &system
	w.RenderSystems = append(w.RenderSystems, owned)
	fmt.Println("Adding Renderer: ", reflect.TypeOf(system))
	(*owned).Init(w)
	fmt.Println("initialized: ", reflect.TypeOf(system))
}

func (w *World) CreateMultipleEntitiesFromJson(jsonStr string) (e []*Entity, er error) {
	var data []json.RawMessage

	err := json.Unmarshal([]byte(jsonStr), &data)

	if err != nil {
		return nil, err
	}

	entities := []*Entity{}

	for i := range data {
		e, err := w.CreateEntityFromJson(string(data[i]))

		if err != nil {
			return nil, err
		}

		entities = append(entities, &e)
	}

	return entities, nil
}

func (w *World) CreateEntityFromJson(jsonStr string) (e Entity, er error) {

	var data map[string][]json.RawMessage

	err := json.Unmarshal([]byte(jsonStr), &data)

	components := data["components"]

	var idJson struct {
		Id string
	}

	err = json.Unmarshal([]byte(jsonStr), &idJson)

	id, err := strconv.ParseInt(idJson.Id, 10, 64)

	entity := Entity{Id: id, Components: make(map[int]Component)}

	for i := range components {

		component := components[i]

		var comp dynamic.Type

		_ = json.Unmarshal(component, &comp)

		v := comp.Value()

		if v != nil {
			value := v.(Component)
			entity.Components[value.Id()] = value
		} else {
			fmt.Println("Entity {", entity.Id, "}", " Component ignored =", string(component))
		}

	}

	return entity, err

}

func (w *World) DoesEntityHaveAllRequiredComponentTypes(entity *Entity, requiredComponents []ComponentType) bool {
	count := 0
	for i := range entity.Components {
		id := entity.Components[i].Id()

		found := false
		for j := range requiredComponents {
			if (int(requiredComponents[j]) == id) {
				found = true
			}
		}

		if found {
			count++
		}
	}

	return count == len(requiredComponents)
}

// queues entity for spawning
func (w *World) Spawn(entity Entity) {
	w.ToSpawn = append(w.ToSpawn, entity)
}

// queues entity for destruction
func (w *World) Destroy(entityId int64) {
	w.ToDestroy = append(w.ToDestroy, entityId)
}

func (w *World) AddEntityToWorld(entity Entity) {

	for i := range entity.Components {
		entity.Components[i].CreateComponent()
	}

	for i := range w.Systems {
		system := *w.Systems[i]

		if w.DoesEntityHaveAllRequiredComponentTypes(&entity, system.RequiredComponentTypes()) {
			system.AddToStorage(&entity)
		}
	}

	for i := range w.RenderSystems {
		system := *w.RenderSystems[i]

		if w.DoesEntityHaveAllRequiredComponentTypes(&entity, system.RequiredComponentTypes()) {
			system.AddToStorage(&entity)
		}
	}

	w.Entities[entity.Id] = &entity
}

func (w *World) FetchAndIncrementId() int64 {
	temp := w.IdIndex

	w.IdIndex++
	return temp
}

func (w *World) AddComponentToEntity(c Component, entity Entity) {
	c.CreateComponent();
	entity.Components[c.Id()] = c

	for i := range w.Systems {
		system := *w.Systems[i]

		if w.DoesEntityHaveAllRequiredComponentTypes(&entity, system.RequiredComponentTypes()) {
			system.AddToStorage(&entity)
		}
	}

	for i := range w.RenderSystems {
		system := *w.RenderSystems[i]

		if w.DoesEntityHaveAllRequiredComponentTypes(&entity, system.RequiredComponentTypes()) {
			system.AddToStorage(&entity)
		}
	}
}

func (w *World) RemoveEntity(id int64) {

	entity := w.Entities[id]

	for i := range w.Systems {
		system := *w.Systems[i]

		if w.DoesEntityHaveAllRequiredComponentTypes(entity, system.RequiredComponentTypes()) {
			system.RemoveFromStorage(entity)
		}
	}

	for i := range w.RenderSystems {
		system := *w.RenderSystems[i]

		if w.DoesEntityHaveAllRequiredComponentTypes(entity, system.RequiredComponentTypes()) {
			system.AddToStorage(entity)
		}
	}

	w.Entities = nil
}

func (w *World) ResetToTick(tick int64) {

	diff := int(w.CurrentTick - tick)

	if (len(w.Cache)- 1) - int(diff) > 0 {

		index := (len(w.Cache) - 1) - diff

		w.Entities = w.Cache[index]
		w.Input = w.CacheInput[index]

		mask := int32(0)
		for i := 0; i < diff; i++ {
			mask = (mask << 1) | 1
		}

		w.ValidatedBuffer = w.ValidatedBuffer | mask

		w.Cache = w.Cache[:index]
		w.CacheInput = w.CacheInput[:index]
	}

}

func (w *World) Resimulate (tick int64) {

	diff := int(w.CurrentTick - tick)

	w.IsResimulating = true
	for i := 0; i < diff; i++ {
		w.Input = w.CacheInput[diff + i]
		w.Update(FIXED_DELTA)
	}
	w.IsResimulating = false

}

func (w *World) CacheState() {
	clone := map[int64]*Entity{}

	for i, entity := range w.Entities {
		clone[i] = entity.Clone()
	}

	inputClone := w.Input.Clone()

	w.Cache = append(w.Cache, clone)
	w.CacheInput = append(w.CacheInput, inputClone)

	w.ValidatedBuffer = w.ValidatedBuffer << 1

	if len(w.Cache) > 32 {
		w.Cache = w.Cache[1:]
		w.CacheInput = w.CacheInput[1:]
	}
}

func (w *World) CompareEntitiesAtTick(tick int64, tempEntity *Entity) (same bool) {
	diff := int(w.CurrentTick - tick)

	if (len(w.Cache)-1) - diff > 0 {
		return w.Cache[(len(w.Cache)-1) - diff][tempEntity.Id].CompareTo(tempEntity)
	}

	return true
}
