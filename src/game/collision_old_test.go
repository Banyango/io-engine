package game

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"io-engine-backend/src/shared"
)

func TestCollisionSystem_CollisionTypes(t *testing.T) {

	system := CollisionSystem{}

	types := system.RequiredComponentTypes()

	assert.Equal(t, shared.PositionComponentType, types[0])
	assert.Equal(t, shared.CollisionComponentType, types[1])

}

func TestCollisionSystem_AddToStorage(t *testing.T) {

	json := `{	
		"id":0, 
		"components":[
			{"Type":"PositionComponent", "Position":[0,1] },
			{"Type":"CollisionComponent", "Size":[0,1] }
		]
	}`

	w := shared.NewWorld()

	system := new(CollisionSystem)

	w.AddSystem(system)
	entity, err := w.CreateEntityFromJson(json)

	entity.Id = w.FetchAndIncrementId()
	w.AddEntityToWorld(entity)

	assert.NoError(t, err)

	assert.NotNil(t, system.positionComponents.Components[entity.Id])
	assert.NotNil(t, system.collisionComponents.Components[entity.Id])

	assert.ObjectsAreEqual(system.positionComponents.Components[entity.Id], entity.Components[int(shared.PositionComponentType)])
	assert.ObjectsAreEqual(system.collisionComponents.Components[entity.Id], entity.Components[int(shared.CollisionComponentType)])

}

func TestCollisionSystem_UpdateSystem(t *testing.T) {

	json := `[
		{	
			"id":0, 
			"components":[
				{"Type":"PositionComponent", "Position":[0,1] },
				{"Type":"CollisionComponent", "Size":[2,2], "Velocity":[1,0] }
			]
		},
		{	
			"id":1, 
			"components":[
				{"Type":"PositionComponent", "Position":[4,1] },
				{"Type":"CollisionComponent", "Size":[2,2], "Velocity":[-1,0] }
			]
		}
	]`

	w := shared.NewWorld()

	system := new(CollisionSystem)

	w.AddSystem(system)
	entities, err := w.CreateMultipleEntitiesFromJson(json)

	assert.NoError(t, err)
	for i := range entities {
		entity := *entities[i]
		entity.Id = w.FetchAndIncrementId()
		w.AddEntityToWorld(entity)
	}

	w.Update(1)
	w.Update(1)
	w.Update(1)

	collider1 := w.Entities[int64(0)].Components[int(shared.CollisionComponentType)].(*CollisionComponent)
	collider2 := w.Entities[int64(1)].Components[int(shared.CollisionComponentType)].(*CollisionComponent)

	assert.True(t, collider1.Right)
	assert.True(t, collider2.Left)

	assert.Equal(t, int64(1), collider1.entitiesCollidingWith[0])
	assert.Equal(t, int64(0), collider2.entitiesCollidingWith[0])

}

