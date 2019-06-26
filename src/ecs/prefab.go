package ecs

import (
	"encoding/json"
	"errors"
	"fmt"
)

type PrefabData struct {
	data GameDataJson
	Prefabs map[int]Entity
}

type GameDataJson struct {
	Name string `json:"name"`
	Version string `json:"version"`
	Globals json.RawMessage `json:"globals"`
	Prefabs map[string]json.RawMessage `json:"prefabs"`
}

func (self *PrefabData) CreatePrefab(id int) (Entity, error) {

	if val, ok := self.Prefabs[id]; ok {

		clone := Entity{}

		clone.Id = val.Id;

		clone.Components = make(map[int]Component)

		for _, comp := range val.Components {
			clone.Components[comp.Id()] = comp.Clone().(Component)
		}

		clone.PrefabId = id

		return clone, nil
	}

	return Entity{}, errors.New("Prefab Doesn't exist")
}

// This will create a new prefab manager.
// Not that systems will need to be manually added by name when the world is created.
func NewPrefabManager (jsonGameData string, world *World) (*PrefabData, error) {

	prefabManager := GameDataJson{}

	err := json.Unmarshal([]byte(jsonGameData), &prefabManager)

	if err != nil {
		fmt.Println("Error with json: ", err)

		return nil, err
	}

	result := PrefabData {
		data:prefabManager,
		Prefabs: map[int]Entity{},
	}

	// Create prefabs
	for i := range prefabManager.Prefabs {
		prefab := prefabManager.Prefabs[i]

		entity, _ := world.CreateEntityFromJson(string(prefab))

		result.Prefabs[int(entity.Id)] = entity
	}

	return &result, nil

}



