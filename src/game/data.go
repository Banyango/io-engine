package game

import (
	"io-engine-backend/src/server"
	. "io-engine-backend/src/shared"
)

type NetworkDataCollectionSystem struct {
	collisionComponents Storage
	positionComponents Storage
}

func (*NetworkDataCollectionSystem) Init() {

}

func (self *NetworkDataCollectionSystem) AddToStorage(entity Entity) {
	for k := range entity.Components {
		component := entity.Components[k].(Component)

		if component.Id() == int(CollisionComponentType) {
			self.collisionComponents.Components[entity.Id] = &component
		} else if component.Id() == int(PositionComponentType) {
			self.positionComponents.Components[entity.Id] = &component
		}
	}
}

func (*NetworkDataCollectionSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{PositionComponentType/*, CollisionComponentType*/}
}

func (self *NetworkDataCollectionSystem) UpdateSystem(delta float64, world *World) {
	networkGlobal := world.Globals[ServerGlobalType].(*server.ServerGlobal)

	for entity, _ := range self.positionComponents.Components {
		position := (*self.positionComponents.Components[entity]).(*PositionComponent)
		collider := (*self.collisionComponents.Components[entity]).(*CollisionComponent)

		networkGlobal.NetworkSendUDP(
			server.NetworkData{
				Position:[]int{position.Position.X(),position.Position.Y()},
				Velocity:[]float32{float32(collider.Velocity.X()), float32(collider.Velocity.Y())},
		})
	}
}

type NetworkDataComponent struct {
	Data server.NetworkData
}

func (*NetworkDataComponent) Id() int {
	return int(NetworkDataComponentType)
}

func (self *NetworkDataComponent) CreateComponent() {

}

func (self *NetworkDataComponent) Clone() Component {
	return new(NetworkDataComponent)
}

func (*NetworkDataComponent) DestroyComponent() {

}

/*
----------------------------------------------------------------------------------------------------------------
Network Client Input
----------------------------------------------------------------------------------------------------------------
*/

type NetworkInfo struct {
	PlayerId uint16
	NetworkInput
}

type NetworkInput struct {
	Up    bool
	Down  bool
	Right bool
	Left  bool
	X     bool
	C     bool
	//Mouse1 bool
	//Mouse2 bool
	//mousePos
}

func (self *NetworkInput) ToBytes() []byte {
	value := byte(0)

	if self.Up {
		setBit(&value, 0)
	}

	if self.Down {
		setBit(&value, 1)
	}

	if self.Right {
		setBit(&value, 2)
	}

	if self.Left {
		setBit(&value, 3)
	}

	if self.X {
		setBit(&value, 4)
	}

	if self.C {
		setBit(&value, 5)
	}

	return []byte{value}
}

func NetworkInputFromBytes(value byte) NetworkInput {
	return NetworkInput{
		Up:    hasBit(value, 0),
		Down:  hasBit(value, 1),
		Right: hasBit(value, 2),
		Left:  hasBit(value, 3),
		X:     hasBit(value, 4),
		C:     hasBit(value, 5),
	}
}

func hasBit(n byte, pos uint) bool {
	val := n & (1 << pos)
	return val > 0
}

func setBit(n *byte, pos uint) {
	*n |= 1 << pos
}
