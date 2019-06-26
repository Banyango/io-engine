package ecs

func AddComponentsToStorage(entity *Entity, storages map[int]*Storage) {
	for i, val := range storages {
		component := entity.Components[i]
		val.Components[entity.Id] = &component
	}
}

func RemoveComponentsFromStorage(entity *Entity, storages map[int]*Storage) {
	id := entity.Id
	for _, val := range storages {
		delete (val.Components,id)
	}
}
