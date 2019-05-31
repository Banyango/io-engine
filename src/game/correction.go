package game

import "io-engine-backend/src/ecs"

type CorrectionSystem struct {
	BufferStorage ecs.Storage
}

func (self *CorrectionSystem) Init() {
	self.BufferStorage = ecs.NewStorage()
}

func (self *CorrectionSystem) AddToStorage(entity *ecs.Entity) {
	storages := map[int]*ecs.Storage{
		int(ecs.BufferComponentType): &self.BufferStorage,
	}
	ecs.AddComponentsToStorage(entity, storages)
}

func (self *CorrectionSystem) RequiredComponentTypes() []ecs.ComponentType {
	return []ecs.ComponentType{ecs.BufferComponentType, ecs.NetworkInstanceComponentType}
}

func (self *CorrectionSystem) UpdateSystem(delta float64, world *ecs.World) {
	//for entity, _ := range self.BufferStorage.Components {

		//buffer := (*self.BufferStorage.Components[entity]).(*BufferComponent)

		//buffer.Buffer[]

	//}
}
