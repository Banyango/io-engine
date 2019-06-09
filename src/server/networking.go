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
	PlayerId uint16
}

// Serialized Bytes of the entity and components.
type NetworkData struct {
	OwnerId   uint16
	NetworkId uint16
	PrefabId  uint16
	Data      map[int][]byte
}

type WorldStatePacket struct {
	Tick      int64
	Destroyed []int
	Created   []NetworkData
	Updates   []NetworkData
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
	Data NetworkData
}

func (*NetworkInstanceComponent) Id() int {
	return int(NetworkInstanceComponentType)
}

func (self *NetworkInstanceComponent) CreateComponent() {
	self.Data.Data = make(map[int][]byte)
}

func (*NetworkInstanceComponent) DestroyComponent() {

}

func (*NetworkInstanceComponent) Clone() Component {
	return new(NetworkInstanceComponent)
}

type NetworkInstanceDataCollectionSystem struct {
	NetworkInstances Storage
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

func (self *NetworkInstanceDataCollectionSystem) RemoveFromStorage(entity *Entity) {
	storages := map[int]*Storage{
		int(NetworkInstanceComponentType): &self.NetworkInstances,
	}
	RemoveComponentsFromStorage(entity, storages)
}

func (self *NetworkInstanceDataCollectionSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{NetworkInstanceComponentType}
}

func (self *NetworkInstanceDataCollectionSystem) UpdateSystem(delta float64, world *World) {

	for entity, _ := range self.NetworkInstances.Components {

		instance := (*self.NetworkInstances.Components[entity]).(*NetworkInstanceComponent)

		entity := world.Entities[entity]

		for _, val := range entity.Components {
			if syncVar, ok := val.(WriteSyncUDP); ok {
				syncVar.WriteUDP(&instance.Data)
			}
		}

	}
}

// todo there is an issue with creating client side entities whereby the id will collide with the server id of other network entities.
