package server

import (
	"github.com/goburrow/dynamic"
	. "io-engine-backend/src/ecs"
)

type PlayerStateSystem struct{
	GameState Storage
	Network Storage
}

func (self *PlayerStateSystem) Init() {
	dynamic.Register("PlayerStateComponent", func() interface{} {
		return &PlayerStateComponent{}
	})

	self.GameState = NewStorage()
	self.Network = NewStorage()
}

func (self *PlayerStateSystem) AddToStorage(entity *Entity) {
	for k := range entity.Components {
		component := entity.Components[k].(Component)

		if component.Id() == int(StateComponentType) {
			self.GameState.Components[entity.Id] = &component
		} else if component.Id() == int(NetworkConnectionComponentType) {
			self.Network.Components[entity.Id] = &component
		}
	}
}

func (*PlayerStateSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{StateComponentType, NetworkConnectionComponentType}
}

func (self *PlayerStateSystem) UpdateSystem(delta float64, world *World) {
	for entity, _ := range self.GameState.Components {

		connected := (*self.Network.Components[entity]).(*NetworkConnectionComponent)
		state := (*self.GameState.Components[entity]).(*PlayerStateComponent)

		if state.State == Connecting && connected.IsDataChannelOpen {
			state.State = Spawn
		}

		if state.State == Spawn {
			global := world.Globals[int(ServerGlobalType)].(*ServerGlobal)

			global.ServerSpawn(connected.PlayerId, world, 1, 2, true)

			state.State = Playing
		}
	}
}

type PlayerState int;

const (
	Connecting PlayerState = iota
	Spawn
	Playing
	Dead
)

type PlayerStateComponent struct {
	State PlayerState
}

func (*PlayerStateComponent) Id() int {
	return int(StateComponentType)
}

func (self *PlayerStateComponent) CreateComponent() {
	self.State = Connecting
}

func (*PlayerStateComponent) DestroyComponent() {

}

func (self *PlayerStateComponent) Clone() Component {
	return new(PlayerStateComponent)
}


