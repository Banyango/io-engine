package ecs_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/Banyango/io-engine/src/game"
	"github.com/Banyango/io-engine/src/math"
	"github.com/Banyango/io-engine/src/ecs"
	"testing"
)

func TestNewPrefabManager(t *testing.T) {

	json := `{
	"name": "test",
	"version": "0.0.1",	
	"globals":[
		{"Type":"RenderGlobal", "CanvasElementId":"mycanvas", "Width":600, "Height":500},
		{"Type":"RawInputGlobal"},
		{"Type":"InputGlobal"}
	],
	"prefabs": {
		"player" : {
			"id": 0,
			"components": [
				{"Type":"PositionComponent", "Position":[0,0] },
				{"Type":"CollisionComponent", "Size":[2,2], "Velocity":[0,0] },
				{"Type":"CircleRendererComponent", "Size":[6,6], "Radius":12.3, "Color":"#001121" },
				{"Type":"ArcadeMovementComponent", "MaxSpeed":[200,200], "Speed":400, "Drag":0.8, "Gravity":[0,4] }
			]
		}
	}
}`

	w := ecs.NewWorld()

	w.AddSystem(new(game.CollisionSystem))
	//w.AddSystem(new(client.InputSystem))
	w.AddSystem(new(game.KeyboardMovementSystem))

	pm, err := ecs.NewPrefabManager(json, w)

	assert.NoError(t, err)
	assert.NotNil(t, pm.Prefabs[0])

}

func TestPrefabClone(t *testing.T) {

	json := `{
	"name": "test",
	"version": "0.0.1",	
	"globals":[
		{"Type":"RenderGlobal", "CanvasElementId":"mycanvas", "Width":600, "Height":500},
		{"Type":"RawInputGlobal"},
		{"Type":"InputGlobal"}
	],
	"prefabs": {
		"player" : {
			"id": 0,
			"components": [
				{"Type":"PositionComponent", "Position":[0,0] },
				{"Type":"CollisionComponent", "Size":[2,2], "Velocity":[0,0] },
				{"Type":"CircleRendererComponent", "Size":[6,6], "Radius":12.3, "Color":"#001121" },
				{"Type":"ArcadeMovementComponent", "MaxSpeed":[200,200], "Speed":400, "Drag":0.8, "Gravity":[0,4] }
			]
		}
	}
}`

	w := ecs.NewWorld()

	w.AddSystem(new(game.CollisionSystem))
	//w.AddSystem(new(client.InputSystem))
	w.AddSystem(new(game.KeyboardMovementSystem))

	pm, err := ecs.NewPrefabManager(json, w)

	assert.NoError(t, err)
	assert.NotNil(t, pm.Prefabs[0])

	entity1, err := pm.CreatePrefab(0)
	assert.NoError(t, err)

	entity2, err := pm.CreatePrefab(0)
	assert.NoError(t, err)

	entity1.Components[0].(*game.PositionComponent).Position = math.NewVectorInt(1,1)

	assert.NotEqual(t, 1, entity2.Components[0].(*game.PositionComponent).Position.X())
	assert.NotEqual(t, 1, entity2.Components[0].(*game.PositionComponent).Position.Y())
}
