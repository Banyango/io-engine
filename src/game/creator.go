package game

import . "io-engine-backend/src/shared"

type CreatorSystem struct {

}

func (CreatorSystem) Init() {

}

func (CreatorSystem) AddToStorage(entity Entity) {

}

func (CreatorSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{}
}

func (CreatorSystem) UpdateSystem(delta float64, world *World) {
	creator := world.Globals[CreatorGlobalType].(*CreatorGlobal)

	for _, entity := range creator.ToCreate {
		entity.Id = world.FetchAndIncrementId()
		world.AddEntityToWorld(entity)
	}

	creator.ToCreate = []Entity{}

	// remove entity.
	//for _, entity := range creator.ToRemove {
	//	world.AddEntityToWorld(entity)
	//}
	//
	//creator.ToRemove = []Entity{}
}

type CreatorGlobal struct {
	PrefabData *PrefabData
	ToCreate   []Entity
	ToRemove   []Entity
}

func (self *CreatorGlobal) Id() int {
	return CreatorGlobalType
}

func (self *CreatorGlobal) CreateGlobal(world *World) {
	self.ToCreate = []Entity{}
	self.PrefabData = world.PrefabData;
}
