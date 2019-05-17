package client

import (
	. "io-engine-backend/src/shared"
)

type ClientNetworkDataSystem struct {
	collisionComponents Storage
	positionComponents  Storage
}

func (*ClientNetworkDataSystem) Init() {

}

func (self *ClientNetworkDataSystem) AddToStorage(entity Entity) {
	for k := range entity.Components {
		component := entity.Components[k].(Component)

		if component.Id() == int(CollisionComponentType) {
			self.collisionComponents.Components[entity.Id] = &component
		} else if component.Id() == int(PositionComponentType) {
			self.positionComponents.Components[entity.Id] = &component
		}
	}
}

func (*ClientNetworkDataSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{PositionComponentType, CollisionComponentType}
}

func (self *ClientNetworkDataSystem) UpdateSystem(delta float64, world *World) {

}





