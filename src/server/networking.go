package server

import (
	. "github.com/Banyango/io-engine/src/ecs"
)

/*
----------------------------------------------------------------------------------------------------------------
Network Shared structs
----------------------------------------------------------------------------------------------------------------
*/

type RoundTripTime struct {
	PlayerId       PlayerId
	RecTime        int64
	SentTimeClient int64
	SentTimeServer int64
}

// Serialized Bytes of the entity and components.
type NetworkData struct {
	OwnerId   PlayerId
	NetworkId uint16
	PrefabId  uint16
	Data      map[int][]byte
}

type ClientWorldStatePacket struct {
	State []byte // gob serialized WorldState struct
	RTT   *RoundTripTime
}

type WorldState struct {
	Tick      int64
	Destroyed []int
	Created   []*NetworkData
	Updates   []*NetworkData
}

type ClientPacket struct {
	Tick  int64
	Time  int64
	Input byte
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
	PrefabId  int
}

func (*NetworkInstanceComponent) Id() int {
	return int(NetworkInstanceComponentType)
}

func (self *NetworkInstanceComponent) CreateComponent() {

}

func (*NetworkInstanceComponent) DestroyComponent() {

}

func (self *NetworkInstanceComponent) Clone() Component {
	component := new(NetworkInstanceComponent)
	component.OwnerId = self.OwnerId
	component.NetworkId = self.NetworkId
	component.PrefabId = self.PrefabId
	return component
}

func (self *NetworkInstanceComponent) Reset(component Component) {

}

/*
----------------------------------------------------------------------------------------------------------------

NetworkInstanceDataCollectionSystem

----------------------------------------------------------------------------------------------------------------
*/

type NetworkInstanceDataCollectionSystem struct {
	NetworkInstances Storage
	Server           *Server
}

func NewNetworkInstanceDataCollectionSystem(server *Server) *NetworkInstanceDataCollectionSystem {
	s := new(NetworkInstanceDataCollectionSystem)
	s.Server = server
	return s
}

func (self *NetworkInstanceDataCollectionSystem) RemoveFromStorage(entity *Entity) {
	keys := map[int]*Storage{
		int(NetworkInstanceComponentType): &self.NetworkInstances,
	}

	RemoveComponentsFromStorage(entity, keys)
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

		entity := world.Entities[entity]

		data := NetworkData{
			OwnerId:   instance.OwnerId,
			NetworkId: instance.NetworkId,
			PrefabId:  uint16(entity.PrefabId),
			Data:      map[int][]byte{},
		}

		for _, val := range entity.Components {
			if syncVar, ok := val.(WriteSyncUDP); ok {
				syncVar.WriteUDP(&data)
			}
		}

		self.Server.CurrentState.Updates = append(self.Server.CurrentState.Updates, &data)
	}
}

/*
----------------------------------------------------------------------------------------------------------------

NetworkInputFutureCollectionSystem

----------------------------------------------------------------------------------------------------------------
*/

type NetworkInputFutureCollectionSystem struct {
}

func (self *NetworkInputFutureCollectionSystem) RemoveFromStorage(entity *Entity) {

}

func (self *NetworkInputFutureCollectionSystem) Init(w *World) {

}

func (self *NetworkInputFutureCollectionSystem) AddToStorage(entity *Entity) {

}

func (self *NetworkInputFutureCollectionSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{}
}

func (self *NetworkInputFutureCollectionSystem) UpdateSystem(delta float64, world *World) {
	for i := range world.Future {
		if world.Future[i].Tick == world.CurrentTick {
			for id, inputBytes := range world.Future[i].Bytes {
				input := world.InputForPlayer(id)
				if input != nil {
					input.InputFromBytes(inputBytes)
				}
			}
		}
	}

	futureCopy := []*BufferedInput{}
	for i := range world.Future {
		if world.Future[i].Tick > world.CurrentTick {
			futureCopy = append(futureCopy, world.Future[i])
		}
	}
	world.Future = futureCopy
}
