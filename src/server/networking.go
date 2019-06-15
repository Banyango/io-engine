package server

import (
	. "io-engine-backend/src/ecs"
)

/*
----------------------------------------------------------------------------------------------------------------
Network Shared structs
----------------------------------------------------------------------------------------------------------------
*/

// this connection must be a goroutine use a goroutine instead of component thing
type ClientInputPacket struct {
	PlayerId uint16
	Input    byte
}

type ServerConnectionHandshakePacket struct {
	PlayerId PlayerId
}

// Serialized Bytes of the entity and components.
type NetworkData struct {
	OwnerId   PlayerId
	NetworkId uint16
	PrefabId  uint16
	Data      map[int][]byte
}

type WorldStatePacket struct {
	Tick      int64
	Destroyed []int
	Created   []*NetworkData
	Updates   []*NetworkData
}

type ReadSyncUDP interface {
	ReadUDP(networkPacket *NetworkData)
}

type WriteSyncUDP interface {
	WriteUDP(networkPacket *NetworkData)
}

type EntityChangeType byte

const (
	InstantiatedEntityChangeType EntityChangeType = iota
	DestroyedEntityChangeType
)

/*
----------------------------------------------------------------------------------------------------------------
Network Instance Component
----------------------------------------------------------------------------------------------------------------
*/

type NetworkInstanceComponent struct {
	OwnerId   PlayerId
	NetworkId uint16
}

func (*NetworkInstanceComponent) Id() int {
	return int(NetworkInstanceComponentType)
}

func (self *NetworkInstanceComponent) CreateComponent() {

}

func (*NetworkInstanceComponent) DestroyComponent() {

}

func (*NetworkInstanceComponent) Clone() Component {
	return new(NetworkInstanceComponent)
}


func (self *NetworkInstanceComponent) Reset(component Component) {
	if val, ok := component.(*NetworkInstanceComponent); ok {
		self.OwnerId = val.OwnerId
	}
}

/*
----------------------------------------------------------------------------------------------------------------

NetworkInstanceDataCollectionSystem

----------------------------------------------------------------------------------------------------------------
*/

type NetworkInstanceDataCollectionSystem struct {
	NetworkInstances Storage
	CurrentState     WorldStatePacket
}

func (self *NetworkInstanceDataCollectionSystem) RemoveFromStorage(entity *Entity) {

}

func (self *NetworkInstanceDataCollectionSystem) Init(w *World) {
	self.NetworkInstances = NewStorage()
}

func (self *NetworkInstanceDataCollectionSystem) AddToStorage(entity *Entity) {

	keys := map[int]*Storage{
		int(NetworkInstanceComponentType): &self.NetworkInstances,
	}

	AddComponentsToStorage(entity, keys)
}

func (self *NetworkInstanceDataCollectionSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{NetworkInstanceComponentType}
}

func (self *NetworkInstanceDataCollectionSystem) UpdateSystem(delta float64, world *World) {

	for entity, _ := range self.NetworkInstances.Components {

		instance := (*self.NetworkInstances.Components[entity]).(*NetworkInstanceComponent)
		instance.NetworkId = uint16(entity)

		entity := world.Entities[entity]

		data := NetworkData{OwnerId: instance.OwnerId, NetworkId: instance.NetworkId, Data: map[int][]byte{}}

		for _, val := range entity.Components {
			if syncVar, ok := val.(WriteSyncUDP); ok {
				syncVar.WriteUDP(&data)
			}
		}

		self.CurrentState.Updates = append(self.CurrentState.Updates, &data)

	}
}
