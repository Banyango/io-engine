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

type Global interface {
	Id() int
	CreateGlobal(world *World)
}

type ComponentType int

const (
	PositionComponentType ComponentType = iota
	CollisionComponentType
	StateComponentType
	CircleComponentType
	ArcadeMovementComponentType
	NetworkConnectionComponentType
	NetworkInstanceComponentType
	NetworkInputComponentType
)

type GlobalType int
const (
	RawInputGlobalType = iota
	NetworkInputGlobalType
	RenderGlobalType
	CreatorGlobalType
	ServerGlobalType
	ClientGlobalType
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

func (entity Entity) Clone() Entity {
	return entity
}

/**
Systems are behaviors

Systems store a list of components
 */
type System interface {
	Init()
	AddToStorage(entity *Entity)
	RequiredComponentTypes() []ComponentType
	UpdateSystem(delta float64, world *World)
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

type World struct {
	IdIndex       int64
	Systems       []*System
	RenderSystems []*System
	Entities      map[int64]*Entity
	Globals       map[int]Global

	TimeElapsed      int64
	LastFrameTime    int64
	CurrentFrameTime int64

	Interval   int64
	PrefabData *PrefabData
	mux sync.Mutex
}



func (w *World) Update(delta float64) {
	for _, v := range w.Systems {
		(*v).UpdateSystem(delta, w)
	}
}

func (self *World) Render() {
	for _, v := range self.RenderSystems {
		(*v).UpdateSystem(-1, self)
	}
}

func NewWorld() *World {
	world := new(World)

	world.Entities = map[int64]*Entity{}
	world.Globals = map[int]Global{}

	world.Interval = 16

	return world
}

func (w *World) AddSystem(system System) {
	owned := &system
	w.Systems = append(w.Systems, owned)
	fmt.Println("Adding System: ", reflect.TypeOf(system))
	(*owned).Init()
	fmt.Println("initialized: ", reflect.TypeOf(system))
}

func (w *World) AddRenderer(system System) {
	owned := &system
	w.RenderSystems = append(w.RenderSystems, owned)
	fmt.Println("Adding Renderer: ", reflect.TypeOf(system))
	(*owned).Init()
	fmt.Println("initialized: ", reflect.TypeOf(system))
}

func (w *World) CreateAndAddGlobalsFromJson(jsonStr string) error {
	var data []json.RawMessage

	err := json.Unmarshal([]byte(jsonStr), &data)

	if err != nil {
		fmt.Println("error: ", err)
		return err
	}

	for i := range data {
		e, err := w.createGlobalFromJson(string(data[i]))

		if e == nil {
			fmt.Println("Global ignored: ", string(data[i]))
			continue;
		}

		fmt.Println("Creating global: ", reflect.TypeOf(e))

		if err != nil {
			return err
		}

		w.addGlobal(e);
	}

	return nil
}

func (w *World) createGlobalFromJson(jsonStr string) (c Global, er error) {

	var comp dynamic.Type

	err := json.Unmarshal([]byte(jsonStr), &comp)

	if err != nil {
		fmt.Println("Error creating global ", jsonStr)
		return nil, nil
	}

	v := comp.Value()

	return v.(Global), nil

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

	var idJson struct{
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
			fmt.Println("Entity {",entity.Id,"}", " Component ignored =", string(component))
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
			if(int(requiredComponents[j]) == id) {
				found = true
			}
		}

		if found {
			count++
		}
	}

	return count == len(requiredComponents)
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
	w.mux.Lock()

	temp := w.IdIndex

	w.IdIndex++

	w.mux.Unlock()
	return temp
}

func (w *World) addGlobal(global Global) {
	w.Globals[global.Id()] = global
	global.CreateGlobal(w)
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
