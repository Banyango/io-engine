package ecs

func AddComponentsToStorage(entity *Entity, storages map[int]*Storage) {
	for i, val := range storages {
		component := entity.Components[i]
		val.Components[entity.Id] = &component
	}
}
