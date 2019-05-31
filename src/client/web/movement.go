package web

import (
	"github.com/goburrow/dynamic"
	. "io-engine-backend/src/ecs"
	. "io-engine-backend/src/game"
	"io-engine-backend/src/math"
	"io-engine-backend/src/server"
)

type ClientMovementSystem struct {
	collisionComponents Storage
	arcadeComponents    Storage
}

func (self *ClientMovementSystem) Init() {
	dynamic.Register("ArcadeMovementComponent", func() interface{} {
		return &ArcadeMovementComponent{}
	})

	self.collisionComponents = NewStorage()
	self.arcadeComponents = NewStorage()
}

func (self *ClientMovementSystem) AddToStorage(entity *Entity) {
	for k := range entity.Components {
		component := entity.Components[k].(Component)

		if component.Id() == int(CollisionComponentType) {
			self.collisionComponents.Components[entity.Id] = &component
		} else if component.Id() == int(ArcadeMovementComponentType) {
			self.arcadeComponents.Components[entity.Id] = &component
		}
	}
}

func (self *ClientMovementSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{CollisionComponentType, ArcadeMovementComponentType}
}

func (self *ClientMovementSystem) UpdateSystem(delta float64, world *World) {
	global := world.Globals[int(RawInputGlobalType)].(*RawInputGlobal)

	for entity, _ := range self.collisionComponents.Components {

		arcade := (*self.arcadeComponents.Components[entity]).(*ArcadeMovementComponent)
		collider := (*self.collisionComponents.Components[entity]).(*CollisionComponent)

		direction := math.NewVector(float64(0), float64(0))

		if global != nil {
			if global.AnyKeyPressed() {

				if global.KeyPressed[server.Up] {
					direction = direction.Add(math.VectorUp())
				}

				if global.KeyPressed[server.Down] {
					direction = direction.Add(math.VectorDown())
				}

				if global.KeyPressed[server.Left] {
					direction = direction.Add(math.VectorRight())
				}

				if global.KeyPressed[server.Right] {
					direction = direction.Add(math.VectorLeft())
				}

				collider.Velocity = collider.Velocity.Add(direction.Scale(arcade.Speed))

			}
		}

		collider.Velocity = collider.Velocity.Add(arcade.Gravity).Scale(arcade.Drag).Clamp(arcade.MaxSpeed.Neg(), arcade.MaxSpeed)
	}
}