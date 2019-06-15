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

func (self *KeyboardMovementSystem) Init(w *World) {

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

func (self *KeyboardMovementSystem) RemoveFromStorage(entity *Entity) {
	storages := map[int]*Storage{
		int(ArcadeMovementComponentType): &self.arcadeComponents,
		int(CollisionComponentType):      &self.collisionComponents,
		int(NetworkInstanceComponentType):      &self.networkInstance,
	}
	RemoveComponentsFromStorage(entity, storages)
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

	for entity, _ := range self.collisionComponents.Components {

		arcade := (*self.arcadeComponents.Components[entity]).(*ArcadeMovementComponent)
		collider := (*self.collisionComponents.Components[entity]).(*CollisionComponent)
		net := (*self.networkInstance.Components[entity]).(*server.NetworkInstanceComponent)

		direction := math.NewVector(float64(0), float64(0))

		input := world.Input.Player[net.OwnerId]

		if input.AnyKeyPressed() {

			if input.KeyPressed[Up] {
				direction = direction.Add(math.VectorUp())
			}

			if input.KeyPressed[Down] {
				direction = direction.Add(math.VectorDown())
			}

			if input.KeyPressed[Left] {
				direction = direction.Add(math.VectorRight())
			}

			if input.KeyPressed[Right] {
				direction = direction.Add(math.VectorLeft())
			}

			collider.Velocity = collider.Velocity.Add(direction.Scale(arcade.Speed))

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

func (self *ArcadeMovementComponent) AreEquals(component Component) bool {
	if val, ok := component.(*ArcadeMovementComponent); ok {
		return val.Speed == self.Speed
	} else {
		return false
	}
}

func (self *ArcadeMovementComponent) Id() int {
	return int(ArcadeMovementComponentType)
}

func (self *ArcadeMovementComponent) CreateComponent() {

}

func (self *ArcadeMovementComponent) DestroyComponent() {

}

func (self *ArcadeMovementComponent) Reset(component Component) {
	if val, ok := component.(*ArcadeMovementComponent); ok {
		self.Drag = val.Drag
		self.Speed = val.Speed
		self.Gravity.Set(val.Gravity.X(), val.Gravity.Y())
		self.MaxSpeed.Set(val.MaxSpeed.X(), val.MaxSpeed.Y())
	}
}

func (self *ArcadeMovementComponent) Clone() Component {
	component := new(ArcadeMovementComponent)
	component.Speed = self.Speed
	component.Gravity = self.Gravity
	component.MaxSpeed = self.MaxSpeed
	component.Drag = self.Drag
	return component
}
