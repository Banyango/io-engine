package game

import (
	"github.com/goburrow/dynamic"
	"io-engine-backend/src/math"
	. "io-engine-backend/src/shared"
)

type KeyboardMovementSystem struct {
	collisionComponents Storage
	arcadeComponents Storage
}

func (self *KeyboardMovementSystem) Init() {

	dynamic.Register("ArcadeMovementComponent", func() interface{} {
		return &ArcadeMovementComponent{}
	})

	self.collisionComponents = NewStorage()
	self.arcadeComponents = NewStorage()
}

func (self *KeyboardMovementSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{CollisionComponentType, ArcadeMovementComponentType}
}

func (self *KeyboardMovementSystem) UpdateFrequency() int {
	return 60
}

func (self *KeyboardMovementSystem) AddToStorage(entity Entity) {
	for k := range entity.Components {
		component := entity.Components[k].(Component)

		if component.Id() == int(CollisionComponentType) {
			self.collisionComponents.Components[entity.Id] = &component
		} else if component.Id() == int(ArcadeMovementComponentType) {
			self.arcadeComponents.Components[entity.Id] = &component
		}
	}
}

func (self *KeyboardMovementSystem) UpdateSystem(delta float64, world *World) {

	input := world.Globals[InputGlobalType].(*InputGlobal)

	for entity, _ := range self.collisionComponents.Components {

		arcade := (*self.arcadeComponents.Components[entity]).(*ArcadeMovementComponent)
		collider := (*self.collisionComponents.Components[entity]).(*CollisionComponent)

		direction := math.NewVector(float64(0),float64(0))

		if(input.AnyKeyPressed()) {

			if(input.KeyPressed[Up]) {
				//fmt.Println("up")
				direction = direction.Add(math.VectorUp())
			}

			if(input.KeyPressed[Down]) {
				//fmt.Println("down")
				direction = direction.Add(math.VectorDown())
			}

			if(input.KeyPressed[Left]) {
				//fmt.Println("left")
				direction = direction.Add(math.VectorRight())
			}

			if(input.KeyPressed[Right]) {
				//fmt.Println("right")
				direction = direction.Add(math.VectorLeft())
			}

			collider.Velocity = collider.Velocity.Add(direction.Scale(arcade.Speed))

		}

		collider.Velocity = collider.Velocity.Add(arcade.Gravity).Scale(arcade.Drag).Clamp(arcade.MaxSpeed.Neg(), arcade.MaxSpeed)
	}
}

type ArcadeMovementComponent struct {
	Speed float64
	Drag float64
	MaxSpeed math.Vector
	Gravity math.Vector
}

func (self *ArcadeMovementComponent) Id() int {
	return int(ArcadeMovementComponentType)
}

func (self *ArcadeMovementComponent) CreateComponent() {

}

func (self *ArcadeMovementComponent) DestroyComponent() {

}