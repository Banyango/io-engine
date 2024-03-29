package client

import (
	"github.com/goburrow/dynamic"
	. "github.com/Banyango/io-engine/src/ecs"
	. "github.com/Banyango/io-engine/src/game"
	"github.com/Banyango/io-engine/src/math"
)

type ClientMovementSystem struct {
	collisionComponents Storage
	arcadeComponents    Storage
}

func (self *ClientMovementSystem) Init(w *World) {
	dynamic.Register("ArcadeMovementComponent", func() interface{} {
		return &ArcadeMovementComponent{}
	})

	self.collisionComponents = NewStorage()
	self.arcadeComponents = NewStorage()
}

func (self *ClientMovementSystem) AddToStorage(entity *Entity) {
	storages := map[int]*Storage{
		int(CollisionComponentType): &self.collisionComponents,
		int(ArcadeMovementComponentType): &self.arcadeComponents,
	}
	AddComponentsToStorage(entity, storages)
}

func (self *ClientMovementSystem) RemoveFromStorage(entity *Entity) {
	storages := map[int]*Storage{
		int(CollisionComponentType): &self.collisionComponents,
		int(ArcadeMovementComponentType): &self.arcadeComponents,
	}
	RemoveComponentsFromStorage(entity, storages)
}


func (self *ClientMovementSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{CollisionComponentType, ArcadeMovementComponentType}
}

func (self *ClientMovementSystem) UpdateSystem(delta float64, world *World) {

	global := world.Input.Player[0]
	for entity, _ := range self.collisionComponents.Components {

		arcade := (*self.arcadeComponents.Components[entity]).(*ArcadeMovementComponent)
		collider := (*self.collisionComponents.Components[entity]).(*CollisionComponent)

		direction := math.NewVector(float64(0), float64(0))

		if global.AnyKeyPressed() {

			if global.KeyPressed[Up] {
				direction = direction.Add(math.VectorUp())
			}

			if global.KeyPressed[Down] {
				direction = direction.Add(math.VectorDown())
			}

			if global.KeyPressed[Left] {
				direction = direction.Add(math.VectorRight())
			}

			if global.KeyPressed[Right] {
				direction = direction.Add(math.VectorLeft())
			}

			collider.Velocity = collider.Velocity.Add(direction.Scale(arcade.Speed))

		}


		collider.Velocity = collider.Velocity.Add(arcade.Gravity).Scale(arcade.Drag).Clamp(arcade.MaxSpeed.Neg(), arcade.MaxSpeed)
	}
}
