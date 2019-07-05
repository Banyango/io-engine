package client

import (
	"encoding/json"
	"github.com/Banyango/io-engine/src/ecs"
	"github.com/Banyango/io-engine/src/server"
	"github.com/thoas/go-funk"
	"math"
	"syscall/js"
	"time"
)

const CLIENT_TICK_LEAD = 3

type Client struct {
	PlayerId ecs.PlayerId
	Ping     float64
}

func (self *Client) HandleWorldStatePacket(packet *server.WorldState, world *ecs.World, networkInstances *ecs.Storage) {

	if packet == nil {
		return
	}

	if packet.Tick > world.CurrentTick{
		self.HandleResync(packet, world)
	}

	if world.CurrentTick-packet.Tick < ecs.MAX_CACHE_SIZE {
		resimulateRequired := false

		world.LastServerTick = packet.Tick

		if len(packet.Created) > 0 || len(packet.Destroyed) > 0 {
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
			//log("Loop {resimulating}")
			if world.CurrentTick-packet.Tick > ecs.MAX_CACHE_SIZE {
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

			for _, comp := range networkInstances.Components {
				if net, ok := (*comp).(*server.NetworkInstanceComponent); ok {
					found := funk.Find(packet.Updates, func(d *server.NetworkData) bool {
						if d.NetworkId == net.NetworkId {
							return true
						}
						return false
					})

					if found == nil {
						id := self.findEntityIdInStorageForNetworkId(networkInstances, net.NetworkId)
						if id != -1 {
							world.RemoveEntity(id)
						}
					}
				}
			}

			self.destroyEntities(packet, world, networkInstances)

			world.Resimulate(packet.Tick)
		}
	}
}

func (self *Client) findEntityIdInStorageForNetworkPacket(NetworkInstances *ecs.Storage, data *server.NetworkData) int64 {
	for i := range NetworkInstances.Components {
		net := (*NetworkInstances.Components[i]).(*server.NetworkInstanceComponent)
		if net.NetworkId == data.NetworkId {
			return i
		}
	}
	return -1
}

func (self *Client) findEntityIdInStorageForNetworkId(NetworkInstances *ecs.Storage, id uint16) int64 {
	for i := range NetworkInstances.Components {
		net := (*NetworkInstances.Components[i]).(*server.NetworkInstanceComponent)
		if net.NetworkId == id {
			return i
		}
	}
	return -1
}

func logJson(s string, obj interface{}) {
	marshal, _ := json.Marshal(obj)
	log(s, string(marshal))
}

func (self *Client) destroyEntities(packet *server.WorldState, world *ecs.World, storage *ecs.Storage) {
	for _, destroyed := range packet.Destroyed {
		id := self.findEntityIdInStorageForNetworkId(storage, uint16(destroyed))
		log("removing entity:", id)
		if id != -1 {
			world.RemoveEntity(id)
		}
	}
}

func (self *Client) createEntities(packet *server.WorldState, world *ecs.World) {
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
	component := server.NetworkInstanceComponent{NetworkId: data.NetworkId, OwnerId: data.OwnerId, PrefabId: int(data.PrefabId)}
	entity.Components[int(ecs.NetworkInstanceComponentType)] = &component
	world.Log.LogJson("entityId", entity)
	world.AddEntityToWorld(entity)
}

func (self *Client) HandleResync(packet *server.WorldState, world *ecs.World) {

	if packet.Tick - world.CurrentTick > ecs.MAX_CACHE_SIZE {

		newTick := packet.Tick + int64(math.Round(world.Ping / ecs.FIXED_DELTA)) + CLIENT_TICK_LEAD

		world.Reset()
		world.SetToTick(newTick)


	} else {
		newTick := packet.Tick + int64(math.Round(world.Ping / ecs.FIXED_DELTA)) + CLIENT_TICK_LEAD

		world.IsResimulating = true
		for i := world.CurrentTick; i <= world.CurrentTick + newTick; i++ {
			world.Update(ecs.FIXED_DELTA)
		}
		world.IsResimulating = false
		world.SetToTick(newTick)
	}

}

func (self *Client) HandleRTT(rtt *server.RoundTripTime) {
	if rtt != nil && rtt.SentTimeServer != 0 {
		self.PlayerId = rtt.PlayerId

		sentTime := time.Unix(0, rtt.RecTime).Sub(time.Unix(0, rtt.SentTimeClient))
		recTime := time.Now().Sub(time.Unix(0, rtt.SentTimeServer))

		self.Ping = sentTime.Seconds() + recTime.Seconds()
	}
}

func log(str ...interface{}) {
	js.Global().Get("console").Call("log", str...)
}
