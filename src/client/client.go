package client

import (
	"encoding/json"
	"github.com/Banyango/io-engine/src/ecs"
	"github.com/Banyango/io-engine/src/server"
	"syscall/js"
)

type Client struct {
	PlayerId  ecs.PlayerId
	StartTick int64
}

func (self *Client) HandleWorldStatePacket(packet *server.WorldStatePacket, world *ecs.World, NetworkInstances *ecs.Storage) {
	if packet != nil && world.CurrentTick - packet.Tick < ecs.MAX_CACHE_SIZE {
		resimulateRequired := false

		if (len(packet.Created) > 0 || len(packet.Destroyed) > 0) && packet.Tick <= world.CurrentTick {
			resimulateRequired = true
		} else {
			for _, update := range packet.Updates {
				isPeer := self.PlayerId != update.OwnerId
				temp := update.DeserializeNewEntity(world, isPeer)
				temp.Id = self.findEntityIdInStorageForNetworkPacket(NetworkInstances, update)
				if temp.Id >= 0 && !world.CompareEntitiesAtTick(packet.Tick, temp) {
					resimulateRequired = true
					break
				}
			}
		}

		if resimulateRequired {
			//log("Loop {resimulating}")
			if world.CurrentTick - packet.Tick > ecs.MAX_CACHE_SIZE {
				log("skipping packet")
				return
			}

			//log("Resetting to tick", packet.Tick, " from:", world.CurrentTick)
			world.ResetToTick(packet.Tick)

			//log("Creating entities...")
			self.createEntities(packet, world)

			//logJson("packet", packet)
			//log("Updating entities...")
			for _, update := range packet.Updates {
				entityId := self.findEntityIdInStorageForNetworkPacket(NetworkInstances, update)

				if entityId != -1 {
					update.UpdateEntity(world.Entities[entityId])
				}
			}

			world.Resimulate(packet.Tick)
		}
	}
}

func (self *Client) findEntityIdInStorageForNetworkPacket(NetworkInstances *ecs.Storage, update *server.NetworkData) int64 {
	entityId := int64(-1)
	for i := range NetworkInstances.Components {
		net := (*NetworkInstances.Components[i]).(*server.NetworkInstanceComponent)

		if net.NetworkId == update.NetworkId {
			entityId = i
		}
	}
	return entityId
}

func logJson(s string, obj interface{}) {
	marshal, _ := json.Marshal(obj)
	log(s, string(marshal))
}

func (self *Client) destroyEntities(packet *server.WorldStatePacket, world *ecs.World) {
	for _, destroyed := range packet.Destroyed {
		world.RemoveEntity(int64(destroyed))
	}
}

func (self *Client) createEntities(packet *server.WorldStatePacket, world *ecs.World)  {
	for i := range packet.Created {
		created := packet.Created[i]
		log("Creating entity ", created.PrefabId, " owner ", int(created.OwnerId), "network id", created.NetworkId)
		isPeer := self.PlayerId != created.OwnerId

		entity := *created.DeserializeNewEntity(world, isPeer)
		entity.Id = world.FetchAndIncrementId()

		component := server.NetworkInstanceComponent{NetworkId:created.NetworkId, OwnerId:created.OwnerId}
		entity.Components[int(ecs.NetworkInstanceComponentType)] = &component

		world.Log.LogJson("entityId", entity)

		world.AddEntityToWorld(entity)
	}
}



func log(str ...interface{}) {
	js.Global().Get("console").Call("log", str...)
}
