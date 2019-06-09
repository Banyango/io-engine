package game

import (
	. "io-engine-backend/src/ecs"
)

type SpawnListener interface {
	EntityWasSpawned(entity *Entity)
	EntityWasDestroyed(entity int64)
}

type SpawnSystem struct{
	// add listener
	Listeners []SpawnListener
}

func (self *SpawnSystem) AddSpawnListener(listener SpawnListener) {
	self.Listeners = append(self.Listeners, listener)
}

func (self *SpawnSystem) RemoveSpawnListener(listener SpawnListener) {

	i := -1

	for j, val := range self.Listeners {
		if val == listener {
			i = j
		}
	}

	if i >= 0 {
		self.Listeners = removeSpawnListener(self.Listeners, i)
	}

}

func removeSpawnListener(slice []SpawnListener, s int) []SpawnListener {
	return append(slice[:s], slice[s+1:]...)
}

func (self *SpawnSystem) Init(w *World) {

}

func (self *SpawnSystem) AddToStorage(entity *Entity) {

}

func (self *SpawnSystem) RemoveFromStorage(entity *Entity) {

}

func (*SpawnSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{StateComponentType}
}

func (self *SpawnSystem) UpdateSystem(delta float64, world *World) {

	if world.ToSpawn != nil {
		for _, entity := range world.ToSpawn {

			world.FetchAndIncrementId()

			world.AddEntityToWorld(entity)

			for _, listener := range self.Listeners {
				listener.EntityWasSpawned(&entity)
			}
		}
		world.ToSpawn = nil
	}

	if world.ToDestroy != nil {

		for _, id := range world.ToDestroy {
			world.RemoveEntity(id)

			for _, listener := range self.Listeners {
				listener.EntityWasDestroyed(id)
			}
		}
		world.ToDestroy = nil
	}
}


