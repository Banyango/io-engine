package ecs_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"io-engine-backend/src/game"
	"io-engine-backend/src/ecs"
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
	ref.Position.Set(1,1)

	assert.Equal(t, 1, (*s2.Components[0]).(*game.PositionComponent).Position.X())
	assert.Equal(t, 1, (*s2.Components[0]).(*game.PositionComponent).Position.Y())
}
