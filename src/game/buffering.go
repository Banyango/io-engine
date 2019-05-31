package game

import "io-engine-backend/src/ecs"

type BufferSystem struct {
	BufferStorage ecs.Storage
}

func (self *BufferSystem) Init() {
	self.BufferStorage = ecs.NewStorage()
}

func (self *BufferSystem) AddToStorage(entity *ecs.Entity) {
	storages := map[int]*ecs.Storage{
		int(ecs.BufferComponentType): &self.BufferStorage,
	}
	ecs.AddComponentsToStorage(entity, storages)
}

func (self *BufferSystem) RequiredComponentTypes() []ecs.ComponentType {
	return []ecs.ComponentType{ecs.BufferComponentType}
}

func (self *BufferSystem) UpdateSystem(delta float64, world *ecs.World) {
	for entity, _ := range self.BufferStorage.Components {

		buffer := (*self.BufferStorage.Components[entity]).(*BufferComponent)

		component := world.Entities[entity].Components[buffer.Type]

		buffer.CopyStateToBuffer(world.CurrentTick, &component)

	}
}

type BufferState struct {
	Tick  int64
	State ecs.Component
}

type BufferComponent struct {
	Type int
	MaxSize int
	Buffer []BufferState
}

func (self *BufferComponent) Id() int {
	return int(ecs.BufferComponentType)
}

func (self *BufferComponent) CreateComponent() {
	self.MaxSize = 8
	self.Buffer = []BufferState{}
}

func (self *BufferComponent) DestroyComponent() {

}

func (self *BufferComponent) Clone() ecs.Component {
	return &BufferComponent{}
}

func (self *BufferComponent) CopyStateToBuffer(tick int64, component *ecs.Component) {

	self.Buffer = append(
		self.Buffer,
		BufferState{
			Tick:  tick,
			State: (*component).Clone(),
		})

	if len(self.Buffer) > self.MaxSize {
		self.Buffer = self.Buffer[1:]
	}

}
