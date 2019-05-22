package game

import (
	"github.com/goburrow/dynamic"
	. "io-engine-backend/src/ecs"
	"io-engine-backend/src/math"
	"io-engine-backend/src/server"
)

type KeyboardMovementSystem struct {
	collisionComponents Storage
	arcadeComponents    Storage
	networkInstance     Storage
}

func (self *KeyboardMovementSystem) Init() {

	dynamic.Register("ArcadeMovementComponent", func() interface{} {
		return &ArcadeMovementComponent{}
	})

	self.collisionComponents = NewStorage()
	self.arcadeComponents = NewStorage()
	self.networkInstance = NewStorage()
}

func (self *KeyboardMovementSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{CollisionComponentType, ArcadeMovementComponentType, NetworkInstanceComponentType}
}

func (self *KeyboardMovementSystem) UpdateFrequency() int {
	return 60
}

func (self *KeyboardMovementSystem) AddToStorage(entity *Entity) {
	for k := range entity.Components {
		component := entity.Components[k].(Component)

		if component.Id() == int(CollisionComponentType) {
			self.collisionComponents.Components[entity.Id] = &component
		} else if component.Id() == int(ArcadeMovementComponentType) {
			self.arcadeComponents.Components[entity.Id] = &component
		} else if component.Id() == int(NetworkInstanceComponentType) {
			self.networkInstance.Components[entity.Id] = &component
		}
	}
}

func (self *KeyboardMovementSystem) UpdateSystem(delta float64, world *World) {

	global := world.Globals[int(NetworkInputGlobalType)].(*server.NetworkInputGlobal)

	for entity, _ := range self.collisionComponents.Components {

		arcade := (*self.arcadeComponents.Components[entity]).(*ArcadeMovementComponent)
		collider := (*self.collisionComponents.Components[entity]).(*CollisionComponent)
		net := (*self.networkInstance.Components[entity]).(*server.NetworkInstanceComponent)

		direction := math.NewVector(float64(0), float64(0))

		input := global.Inputs[server.NetworkId(net.Data.OwnerId)]

		if input != nil {
			if input.AnyKeyPressed() {

				if input.KeyPressed[server.Up] {
					direction = direction.Add(math.VectorUp())
				}

				if input.KeyPressed[server.Down] {
					direction = direction.Add(math.VectorDown())
				}

				if input.KeyPressed[server.Left] {
					direction = direction.Add(math.VectorRight())
				}

				if input.KeyPressed[server.Right] {
					direction = direction.Add(math.VectorLeft())
				}

				collider.Velocity = collider.Velocity.Add(direction.Scale(arcade.Speed))

			}
		}

		collider.Velocity = collider.Velocity.Add(arcade.Gravity).Scale(arcade.Drag).Clamp(arcade.MaxSpeed.Neg(), arcade.MaxSpeed)
	}
}

type ArcadeMovementComponent struct {
	Speed    float64
	Drag     float64
	MaxSpeed math.Vector
	Gravity  math.Vector
}

func (self *ArcadeMovementComponent) Id() int {
	return int(ArcadeMovementComponentType)
}

func (self *ArcadeMovementComponent) CreateComponent() {

}

func (self *ArcadeMovementComponent) DestroyComponent() {

}

func (self *ArcadeMovementComponent) Clone() Component {
	component := new(ArcadeMovementComponent)
	component.Speed = self.Speed
	component.Gravity = self.Gravity
	component.MaxSpeed = self.MaxSpeed
	component.Drag = self.Drag
	return component
}
