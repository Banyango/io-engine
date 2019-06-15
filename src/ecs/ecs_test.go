package ecs_test

import (
	"github.com/stretchr/testify/assert"
	"io-engine-backend/src/ecs"
	"io-engine-backend/src/game"
	"io-engine-backend/src/math"
	"testing"
)

func TestCreateEntityFromJson(t *testing.T) {

	json := `{	
		"id":0, 
		"components":[
			{"Type":"PositionComponent", "Position":[0,1] },
			{"Type":"CollisionComponent", "Size":[0,1], "Velocity":[0,0] }
		]
	}`

	w := ecs.NewWorld()

	w.AddSystem(new(game.CollisionSystem))

	entity, err := w.CreateEntityFromJson(json)

	assert.NoError(t, err)
	assert.NotNil(t, entity.Components[0])

}

func TestCreateMultipleEntitiesFromJson(t *testing.T) {

	json := `[
		{	
			"id":0, 
			"components":[
				{"Type":"PositionComponent", "Position":[0,1] },
				{"Type":"CollisionComponent", "Size":[0,1] }
			]
		},
		{	
			"id":1, 
			"components":[
				{"Type":"PositionComponent", "Position":[0,1] }
			]
		}
	]`

	w := ecs.World{}

	w.AddSystem(new(game.CollisionSystem))

	entities, err := w.CreateMultipleEntitiesFromJson(json)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(entities))
	assert.Equal(t, int64(0), entities[0].Id)
	assert.Equal(t, int64(1), entities[1].Id)

	assert.Equal(t, 2, len(entities[0].Components))
	assert.Equal(t, 1, len(entities[1].Components))

}

func TestWorld_AddEntityToWorld(t *testing.T) {

	json := `[
		{	
			"id":0, 
			"components":[
				{"Type":"PositionComponent", "Position":[0,1] },
				{"Type":"CollisionComponent", "Size":[0,1] }
			]
		},
		{	
			"id":1, 
			"components":[
				{"Type":"PositionComponent", "Position":[0,1] }
			]
		}
	]`

	w := ecs.World{}

	system := new(game.CollisionSystem)
	w.AddSystem(system)

	entities, err := w.CreateMultipleEntitiesFromJson(json)

	assert.NoError(t, err)

	for i := range entities {
		entity := *entities[i]
		entity.Id = w.FetchAndIncrementId()
		w.AddEntityToWorld(entity)
	}

	// todo this needs to test that storage only has one entity
	//assert.Equal(t,system.)
}

func TestStorage(t *testing.T) {

	entity := ecs.Entity{}
	entity.Components = make(map[int]ecs.Component)
	entity.Components[0] = new(game.PositionComponent)

	s1 := ecs.NewStorage()
	s2 := ecs.NewStorage()

	component := entity.Components[0].(ecs.Component)
	s1.Components[0] = &component

	component2 := entity.Components[0].(ecs.Component)
	s2.Components[0] = &component2

	ref := (*s1.Components[0]).(*game.PositionComponent)
	ref.Position.Set(1, 1)

	assert.Equal(t, 1, (*s2.Components[0]).(*game.PositionComponent).Position.X())
	assert.Equal(t, 1, (*s2.Components[0]).(*game.PositionComponent).Position.Y())
}

func TestWorld_CacheState_AddEntity(t *testing.T) {
	world := ecs.NewWorld()

	entity := ecs.NewEntity()

	entity.Id = world.FetchAndIncrementId()
	entity.Components[int(ecs.PositionComponentType)] = &game.PositionComponent{Position: math.NewVectorInt(1, 1)}

	world.AddEntityToWorld(entity)

	world.CacheState()

	comp := entity.Components[int(ecs.PositionComponentType)].(*game.PositionComponent)
	comp.Position.Set(2, 2)

	assert.Equal(t, 1, len(world.Cache))

	cachedComponent := world.Cache[0][0].Components[int(ecs.PositionComponentType)].(*game.PositionComponent)
	assert.Equal(t, 1, cachedComponent.Position.X())
	assert.Equal(t, 1, cachedComponent.Position.Y())

}

func TestWorld_CacheInput(t *testing.T) {

	world := ecs.NewWorld()

	world.Update(0.016)

	world.Input.Player[0].KeyPressed[ecs.Down] = true

	world.Update(0.016)
	world.Update(0.016)
	world.Update(0.016)

	world.Input.Player[0].KeyPressed[ecs.Down] = false

	world.Update(0.016)

	assert.False(t, world.CacheInput[0].Player[0].KeyPressed[ecs.Down])
	assert.True(t, world.CacheInput[1].Player[0].KeyPressed[ecs.Down])
	assert.True(t, world.CacheInput[2].Player[0].KeyPressed[ecs.Down])
	assert.True(t, world.CacheInput[3].Player[0].KeyPressed[ecs.Down])
	assert.False(t, world.CacheInput[4].Player[0].KeyPressed[ecs.Down])

}

func TestWorld_CacheInputWithReset(t *testing.T) {

	world := ecs.NewWorld()

	world.Update(0.016)

	world.Input.Player[0].KeyPressed[ecs.Down] = true

	world.Update(0.016)
	world.Update(0.016)
	world.Update(0.016)

	world.Input.Player[0].KeyPressed[ecs.Down] = false

	world.Update(0.016)

	world.ResetToTick(1)
	world.Resimulate(1)

	assert.False(t, world.CacheInput[0].Player[0].KeyPressed[ecs.Down])
	assert.True(t, world.CacheInput[1].Player[0].KeyPressed[ecs.Down])
	assert.True(t, world.CacheInput[2].Player[0].KeyPressed[ecs.Down])
	assert.True(t, world.CacheInput[3].Player[0].KeyPressed[ecs.Down])
	assert.False(t, world.CacheInput[4].Player[0].KeyPressed[ecs.Down])

}

func TestWorld_CacheState_Reset(t *testing.T) {
	world := ecs.NewWorld()

	entity := ecs.NewEntity()

	entity.Id = world.FetchAndIncrementId()
	entity.Components[int(ecs.PositionComponentType)] = &game.PositionComponent{Position: math.NewVectorInt(1, 1)}

	world.AddEntityToWorld(entity)

	for i := 0; i < 4; i++ {
		comp := entity.Components[int(ecs.PositionComponentType)].(*game.PositionComponent)
		comp.Position.Set(2*i, 2*i)
		world.Update(0.016)
	}

	world.ResetToTick(1)

	cachedComponent := world.Entities[entity.Id].Components[int(ecs.PositionComponentType)].(*game.PositionComponent)
	assert.Equal(t, 2, cachedComponent.Position.X())
	assert.Equal(t, 2, cachedComponent.Position.Y())
	assert.Equal(t, int32(7), world.ValidatedBuffer)

}

func TestWorld_CacheState_MaxSize(t *testing.T) {
	world := ecs.NewWorld()

	entity := ecs.NewEntity()

	entity.Id = world.FetchAndIncrementId()
	entity.Components[int(ecs.PositionComponentType)] = &game.PositionComponent{Position: math.NewVectorInt(1, 1)}

	world.AddEntityToWorld(entity)

	for i := 0; i < 32; i++ {
		world.Update(0.016)
	}

	world.ResetToTick(5)

	for i := 0; i < 32; i++ {
		world.Update(0.016)
	}

	assert.Equal(t, 32, len(world.Cache))
	assert.Equal(t, 32, len(world.CacheInput))
	assert.Equal(t, int32(0), world.ValidatedBuffer)

}

// test system for simulation
type testSystem struct {
}

func (*testSystem) Init(w *ecs.World)                    {}
func (*testSystem) AddToStorage(entity *ecs.Entity)      {}
func (*testSystem) RemoveFromStorage(entity *ecs.Entity) {}

func (*testSystem) RequiredComponentTypes() []ecs.ComponentType {
	return []ecs.ComponentType{ecs.PositionComponentType}
}

func (*testSystem) UpdateSystem(delta float64, world *ecs.World) {
	for i := range world.Entities {
		comp := world.Entities[i].Components[int(ecs.PositionComponentType)].(*game.PositionComponent)
		comp.Position = comp.Position.Add(math.NewVectorInt(1, 1))
	}
}

func TestWorld_Resimulate(t *testing.T) {
	world := ecs.NewWorld()

	world.AddSystem(new(testSystem))

	entity := ecs.NewEntity()

	entity.Id = world.FetchAndIncrementId()
	entity.Components[int(ecs.PositionComponentType)] = &game.PositionComponent{Position: math.NewVectorInt(0, 0)}

	world.AddEntityToWorld(entity)

	for i := 0; i < 32; i++ {
		world.Update(0.016)
	}

	{
		cachedComponent := entity.Components[int(ecs.PositionComponentType)].(*game.PositionComponent)
		assert.Equal(t, 32, cachedComponent.Position.X())
		assert.Equal(t, 32, cachedComponent.Position.Y())
	}

	serverState := ecs.NewEntity()
	serverState.Id = entity.Id
	serverState.Components[int(ecs.PositionComponentType)] = &game.PositionComponent{Position: math.NewVectorInt(7, 7)}

	same := world.CompareEntitiesAtTick(2, &serverState)

	assert.False(t, same)

	world.ResetToTick(2)

	serverResetEntity := world.Entities[serverState.Id].Components[int(ecs.PositionComponentType)].(*game.PositionComponent)
	serverResetEntity.Position = math.NewVectorInt(7, 7)

	world.Resimulate(2)

	cachedComponent := world.Entities[serverState.Id].Components[int(ecs.PositionComponentType)].(*game.PositionComponent)
	assert.Equal(t, 37, cachedComponent.Position.X())
	assert.Equal(t, 37, cachedComponent.Position.Y())
}
