package server

import (
	"bytes"
	"encoding/gob"
	"fmt"
	. "io-engine-backend/src/ecs"
)

func DecodeNetworkDataBytes(networkPacket *NetworkData, id int, buf interface{}) {
	reader := bytes.NewReader(networkPacket.Data[id])
	enc := gob.NewDecoder(reader)
	err := enc.Decode(buf)
	if err != nil {
		fmt.Println("Error Decoding id:", id, " owner:", networkPacket.OwnerId)
	}
}

func EncodeNetworkDataBytes(networkPacket *NetworkData, id int, buf interface{}) {
	buffer := bytes.Buffer{}
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(buf)
	if err != nil {
		fmt.Println("Error encoding id:", id, " owner:", networkPacket.OwnerId, err)
	}
	networkPacket.Data[id] = buffer.Bytes()
}

func ServerSpawn(ownerId uint16, world *World, prefabId byte) {

	entity, err := world.PrefabData.CreatePrefab(int(prefabId))

	if err != nil {
		fmt.Println("Error Spawning entity: ", int(prefabId))
		return
	}

	entity.Id = world.FetchAndIncrementId()

	networkInstanceComponent := new(NetworkInstanceComponent)

	networkInstanceComponent.Data = NetworkData{}
	networkInstanceComponent.Data.OwnerId = ownerId
	networkInstanceComponent.Data.NetworkId = uint16(entity.Id)

	entity.Components[int(NetworkInstanceComponentType)] = networkInstanceComponent

	world.Spawn(entity)

}
