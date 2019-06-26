package client

import (
	"encoding/json"
	"github.com/Banyango/io-engine/src/ecs"
	"github.com/Banyango/io-engine/src/server"
	"syscall/js"
)

const CLIENT_TICK_LEAD = 3

type Client struct {
	PlayerId  ecs.PlayerId
	StartTick int64
}

func (self *Client) HandleWorldStatePacket(packet *server.WorldStatePacket, world *ecs.World, networkInstances *ecs.Storage) {

	if packet == nil || packet.Tick > world.CurrentTick {
		return
	}

	if world.CurrentTick - packet.Tick < ecs.MAX_CACHE_SIZE {
		resimulateRequired := false

		world.LastServerTick = packet.Tick

		if (len(packet.Created) > 0 || len(packet.Destroyed) > 0) && packet.Tick <= world.CurrentTick {
			resimulateRequired = true
		} else {
			for _, update := range packet.Updates {
				isPeer := self.PlayerId != update.OwnerId
				temp := update.DeserializeNewEntity(world, isPeer)
				temp.Id = self.findEntityIdInStorageForNetworkPacket(networkInstances, update)
				if temp.Id == -1 || (temp.Id >= 0 && !world.CompareEntitiesAtTick(packet.Tick, temp)) {
					resimulateRequired = true
					break
				}
			}
		}

		if resimulateRequired {
			log("Loop {resimulating}")
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
				entityId := self.findEntityIdInStorageForNetworkPacket(networkInstances, update)

				if entityId != -1 {
					update.UpdateEntity(world.Entities[entityId])
				} else {
					self.createEntity(update, world)
				}
			}

			world.Resimulate(packet.Tick)
		}
	}
}

func (self *Client) findEntityIdInStorageForNetworkPacket(NetworkInstances *ecs.Storage, update *server.NetworkData) int64 {
	for i := range NetworkInstances.Components {
		net := (*NetworkInstances.Components[i]).(*server.NetworkInstanceComponent)
		if net.NetworkId == update.NetworkId {
			return i
		}
	}
	return -1
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
		self.createEntity(created, world)
	}
}

func (self *Client) createEntity(data *server.NetworkData, world *ecs.World) {
	log("Creating entity ", data.PrefabId, " owner ", int(data.OwnerId), "network id", data.NetworkId)
	isPeer := self.PlayerId != data.OwnerId
	entity := *data.DeserializeNewEntity(world, isPeer)
	entity.Id = world.FetchAndIncrementId()
	component := server.NetworkInstanceComponent{NetworkId: data.NetworkId, OwnerId: data.OwnerId}
	entity.Components[int(ecs.NetworkInstanceComponentType)] = &component
	world.Log.LogJson("entityId", entity)
	world.AddEntityToWorld(entity)
}

func (self *Client) HandleHandshake(packet server.ServerConnectionHandshakePacket, world *ecs.World) {
	self.PlayerId = packet.PlayerId
	startTick := packet.ServerTick + CLIENT_TICK_LEAD
	self.StartTick = startTick
	world.SetToTick(startTick)
	//
	//for _, state := range packet.State {
	//	self.createEntity(state, world)
	//}
}

func log(str ...interface{}) {
	js.Global().Get("console").Call("log", str...)
}
