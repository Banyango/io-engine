package client

import (
	"github.com/stretchr/testify/assert"
	"io-engine-backend/src/ecs"
	"io-engine-backend/src/game"
	"io-engine-backend/src/math"
	"io-engine-backend/src/server"
	"io/ioutil"
	"testing"
)

func TestClient_HandleWorldStatePacketNulls(t *testing.T) {

	client := Client{PlayerId: 0}
	world := ecs.NewWorld()
	packet := server.WorldStatePacket{}

	client.HandleWorldStatePacket(&packet, world)

	packet.Created = []*server.NetworkData{}
	packet.Updates = []*server.NetworkData{}

	client.HandleWorldStatePacket(&packet, world)

}

type testSystem struct {
	pos ecs.Storage
}

func (self *testSystem) Init(w *ecs.World) {
	self.pos = ecs.NewStorage()
}

func (self *testSystem) AddToStorage(entity *ecs.Entity) {
	keys := map[int]*ecs.Storage{
		int(ecs.PositionComponentType): &self.pos,
	}
	ecs.AddComponentsToStorage(entity, keys)
}

func (*testSystem) RemoveFromStorage(entity *ecs.Entity) {}

func (*testSystem) RequiredComponentTypes() []ecs.ComponentType {
	return []ecs.ComponentType{ecs.PositionComponentType}
}

func (self *testSystem) UpdateSystem(delta float64, world *ecs.World) {
	global := world.Input.Player[0]

	for entity, _ := range self.pos.Components {

		p := (*self.pos.Components[entity]).(*game.PositionComponent)

		if global.AnyKeyPressed() {
			if global.KeyPressed[ecs.Down] {
				p.Position = p.Position.Add(math.NewVectorInt(1, 1))
			}
		}
	}
}

func createWorld() (*ecs.World, *Client) {
	gameJson, _ := ioutil.ReadFile("../../game.json");
	world := ecs.NewWorld()
	client := Client{PlayerId: 0}

	collision := new(game.CollisionSystem)
	movement := new(testSystem)

	world.AddSystem(collision)
	world.AddSystem(movement)

	pm, _ := ecs.NewPrefabManager(string(gameJson), world)

	world.PrefabData = pm

	return world, &client
}

func TestRunClient(t *testing.T) {

	world, client := createWorld()

	data := server.NetworkData{OwnerId: 0, NetworkId: 0, Data: map[int][]byte{}}
	component := game.PositionComponent{Position: math.NewVectorInt(2, 2)}
	component.WriteUDP(&data)

	packet := server.WorldStatePacket{}
	packet.Created = append(packet.Created, &data)

	client.HandleWorldStatePacket(&packet, world)

	assert.Equal(t, 1, len(world.Entities))

	position := world.Entities[0].Components[int(ecs.PositionComponentType)].(*game.PositionComponent)

	assert.Equal(t, 2, position.Position.X())
	assert.Equal(t, 2, position.Position.Y())

}

func TestRunClientUpdateAndCreate(t *testing.T) {

	world, client := createWorld()

	// create
	data := server.NetworkData{OwnerId: 0, NetworkId: 0, Data: map[int][]byte{}}
	component := game.PositionComponent{Position: math.NewVectorInt(2, 2)}
	component.WriteUDP(&data)

	// update
	data2 := server.NetworkData{OwnerId: 0, NetworkId: 0, Data: map[int][]byte{}}
	component2 := game.PositionComponent{Position: math.NewVectorInt(5, 5)}
	component2.WriteUDP(&data2)

	packet := server.WorldStatePacket{}
	packet.Created = append(packet.Created, &data)
	packet.Updates = append(packet.Updates, &data2)

	client.HandleWorldStatePacket(&packet, world)

	assert.Equal(t, 1, len(world.Entities))

	position := world.Entities[0].Components[int(ecs.PositionComponentType)].(*game.PositionComponent)

	assert.Equal(t, 5, position.Position.X())
	assert.Equal(t, 5, position.Position.Y())

}

func TestRunClientUpdateAndCreateWithCache(t *testing.T) {

	world, client := createWorld()

	// create
	data := server.NetworkData{OwnerId: 0, NetworkId: 0, Data: map[int][]byte{}}
	component := game.PositionComponent{Position: math.NewVectorInt(2, 2)}
	component.WriteUDP(&data)

	// update
	data2 := server.NetworkData{OwnerId: 0, NetworkId: 0, Data: map[int][]byte{}}
	component2 := game.PositionComponent{Position: math.NewVectorInt(5, 5)}
	component2.WriteUDP(&data2)

	packet := server.WorldStatePacket{}
	packet.Created = append(packet.Created, &data)
	packet.Updates = append(packet.Updates, &data2)

	world.Update(0.016)
	world.Update(0.016)
	world.Update(0.016)
	world.Update(0.016)
	world.Update(0.016)

	client.HandleWorldStatePacket(&packet, world)

	assert.Equal(t, 1, len(world.Entities))

	position := world.Entities[0].Components[int(ecs.PositionComponentType)].(*game.PositionComponent)

	assert.Equal(t, 5, position.Position.X())
	assert.Equal(t, 5, position.Position.Y())

}

func TestRunClientUpdateAndCreateWithCacheAndInput(t *testing.T) {

	world, client := createWorld()

	// create
	data := server.NetworkData{OwnerId: 0, NetworkId: 0, PrefabId: 0, Data: map[int][]byte{}}
	component := game.PositionComponent{Position: math.NewVectorInt(0, 0)}
	component.WriteUDP(&data)

	packet := server.WorldStatePacket{Tick: 0}
	packet.Created = append(packet.Created, &data)

	world.Update(0.016)

	client.HandleWorldStatePacket(&packet, world)

	world.Input.Player[0].KeyPressed[ecs.Down] = true

	world.Update(0.016)
	world.Update(0.016)
	world.Update(0.016)

	world.Input.Player[0].KeyPressed[ecs.Down] = false

	world.Update(0.016)

	position := world.Entities[0].Components[int(ecs.PositionComponentType)].(*game.PositionComponent)
	assert.Equal(t, 3, position.Position.X())
	assert.Equal(t, 3, position.Position.Y())

	// update
	data2 := server.NetworkData{OwnerId: 0, NetworkId: 0, Data: map[int][]byte{}}
	component2 := game.PositionComponent{Position: math.NewVectorInt(-3, -3)}
	component2.WriteUDP(&data2)

	packet2 := server.WorldStatePacket{Tick: 1}
	packet2.Updates = append(packet.Updates, &data2)

	client.HandleWorldStatePacket(&packet2, world)

	position1 := world.Entities[0].Components[int(ecs.PositionComponentType)].(*game.PositionComponent)
	assert.Equal(t, 0, position1.Position.X())
	assert.Equal(t, 0, position1.Position.Y())

}
