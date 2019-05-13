package shared_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"io-engine-backend/src/game"
	"io-engine-backend/src/shared"
)

func TestCreateEntityFromJson(t *testing.T) {

	json := `{	
		"id":0, 
		"components":[
			{"Type":"PositionComponent", "Position":[0,1] },
			{"Type":"CollisionComponent", "Size":[0,1] }
		]
	}`

	w := shared.NewWorld()

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

	w := shared.World{}

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

	w := shared.World{}

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

func TestCreateGlobalFromJson(t *testing.T) {

	globals := `[
		{"Type":"RenderGlobal", "CanvasElementId":"mycanvas"},
		{"Type":"RawInputGlobal"}
	]`

	w := shared.NewWorld()

	w.AddSystem(new(game.CollisionSystem))
	//w.AddSystem(new(game.CanvasRenderSystem))
	//w.AddSystem(new(game.InputSystem))

	err := w.CreateAndAddGlobalsFromJson(globals)

	assert.NoError(t, err)
	assert.NotNil(t, w.Globals)
	assert.NotNil(t, 2, len(w.Globals))

}
