package client

import (
	log2 "github.com/labstack/gommon/log"
	"io-engine-backend/src/ecs"
	"io-engine-backend/src/server"
)

type Client struct {
	PlayerId  ecs.PlayerId
	StartTick int64
}

func (self *Client) HandleWorldStatePacket(packet *server.WorldStatePacket, world *ecs.World) {
	if packet != nil && world.CurrentTick - packet.Tick < ecs.MAX_CACHE_SIZE {
		resimulateRequired := false

		if (len(packet.Created) > 0 || len(packet.Destroyed) > 0) && packet.Tick <= world.CurrentTick {
			resimulateRequired = true
		} else {
			for _, update := range packet.Updates {
				isPeer := self.PlayerId != update.OwnerId
				temp := update.DeserializeNewEntity(world, isPeer)
				temp.Id = int64(update.NetworkId)
				if !world.CompareEntitiesAtTick(packet.Tick, temp) {
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

			world.ResetToTick(packet.Tick)

			self.createEntities(packet, world)

			for _, update := range packet.Updates {
				update.UpdateEntity(world.Entities[int64(update.NetworkId)])
			}

			world.Resimulate(packet.Tick)
		}
	}
}

func (self *Client) destroyEntities(packet *server.WorldStatePacket, world *ecs.World) {
	for _, destroyed := range packet.Destroyed {
		world.RemoveEntity(int64(destroyed))
	}
}

func (self *Client) createEntities(packet *server.WorldStatePacket, world *ecs.World)  {
	for _, created := range packet.Created {
		log("Creating entity ", created.PrefabId, " owner ", created.OwnerId)
		isPeer := self.PlayerId != created.OwnerId

		entity := *created.DeserializeNewEntity(world, isPeer)
		entity.Id = int64(created.NetworkId)

		world.AddEntityToWorld(entity)
	}
}

func log(s... interface{}) {
	log2.Info(s...)
}