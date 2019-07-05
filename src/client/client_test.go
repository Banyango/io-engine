package client

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"github.com/Banyango/io-engine/src/ecs"
	"github.com/Banyango/io-engine/src/game"
	"github.com/Banyango/io-engine/src/math"
	"github.com/Banyango/io-engine/src/server"
	"io/ioutil"
	"testing"
)

func TestClient_HandleWorldStatePacketNulls(t *testing.T) {

	client := Client{PlayerId: 0}
	world := ecs.NewWorld()
	packet := server.WorldState{}

	client.HandleWorldStatePacket(&packet, world, nil)

	packet.Created = []*server.NetworkData{}
	packet.Updates = []*server.NetworkData{}

	client.HandleWorldStatePacket(&packet, world, nil)

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

type testNetworkSystem struct {
	store ecs.Storage
}

func (self *testNetworkSystem) Init(w *ecs.World) {
	self.store = ecs.NewStorage()
}

func (self *testNetworkSystem) AddToStorage(entity *ecs.Entity) {
	ecs.AddComponentsToStorage(entity, map[int]*ecs.Storage{int(ecs.NetworkInstanceComponentType):&self.store})
}

func (self *testNetworkSystem) RequiredComponentTypes() []ecs.ComponentType {
	return []ecs.ComponentType{ecs.NetworkInstanceComponentType}
}

func (self *testNetworkSystem) UpdateSystem(delta float64, world *ecs.World) {

}

func (self *testNetworkSystem) RemoveFromStorage(entity *ecs.Entity) {

}

func createWorld() (*ecs.World, *Client, *ecs.Storage) {
	gameJson, _ := ioutil.ReadFile("../../game.json");
	world := ecs.NewWorld()
	client := Client{PlayerId: 0}
	storage := ecs.NewStorage()

	collision := new(game.CollisionSystem)
	movement := new(testSystem)

	world.AddSystem(collision)
	world.AddSystem(movement)

	pm, _ := ecs.NewPrefabManager(string(gameJson), world)

	world.PrefabData = pm

	return world, &client, &storage
}

func TestRunClient(t *testing.T) {

	world, client, storage := createWorld()

	data := server.NetworkData{OwnerId: 0, NetworkId: 0, Data: map[int][]byte{}}
	component := game.PositionComponent{Position: math.NewVectorInt(2, 2)}
	component.WriteUDP(&data)

	packet := server.WorldState{}
	packet.Created = append(packet.Created, &data)

	client.HandleWorldStatePacket(&packet, world, storage)

	assert.Equal(t, 1, len(world.Entities))

	position := world.Entities[0].Components[int(ecs.PositionComponentType)].(*game.PositionComponent)

	assert.Equal(t, 2, position.Position.X())
	assert.Equal(t, 2, position.Position.Y())

}

func TestRunClientUpdateAndCreate(t *testing.T) {

	world, client, _ := createWorld()

	networkSystem := new(testNetworkSystem)
	world.AddSystem(networkSystem)

	// create
	data := server.NetworkData{OwnerId: 0, NetworkId: 0, Data: map[int][]byte{}}
	component := game.PositionComponent{Position: math.NewVectorInt(2, 2)}
	component.WriteUDP(&data)

	// update
	data2 := server.NetworkData{OwnerId: 0, NetworkId: 0, Data: map[int][]byte{}}
	component2 := game.PositionComponent{Position: math.NewVectorInt(5, 5)}
	component2.WriteUDP(&data2)

	packet := server.WorldState{}
	packet.Created = append(packet.Created, &data)
	packet.Updates = append(packet.Updates, &data2)

	client.HandleWorldStatePacket(&packet, world, &networkSystem.store)

	assert.Equal(t, 1, len(world.Entities))

	position := world.Entities[0].Components[int(ecs.PositionComponentType)].(*game.PositionComponent)

	assert.Equal(t, 5, position.Position.X())
	assert.Equal(t, 5, position.Position.Y())

}

func TestRunClientUpdateBeforeCreate(t *testing.T) {

	world, client, storage := createWorld()

	// update
	data2 := server.NetworkData{OwnerId: 0, NetworkId: 0, Data: map[int][]byte{}}
	component2 := game.PositionComponent{Position: math.NewVectorInt(5, 5)}
	component2.WriteUDP(&data2)

	packet := server.WorldState{}
	packet.Updates = append(packet.Updates, &data2)

	client.HandleWorldStatePacket(&packet, world, storage)

	assert.Equal(t, 0, len(world.Entities))

}

func TestRunClientUpdateAndCreateWithCache(t *testing.T) {

	world, client, storage := createWorld()

	// create
	data := server.NetworkData{OwnerId: 0, NetworkId: 0, Data: map[int][]byte{}}
	component := game.PositionComponent{Position: math.NewVectorInt(2, 2)}
	component.WriteUDP(&data)

	assert.Equal(t, 1, len(world.Entities))

	ecs.AddComponentsToStorage(world.Entities[0], map[int]*ecs.Storage{int(ecs.NetworkInstanceComponentType):storage})

	// update
	data2 := server.NetworkData{OwnerId: 0, NetworkId: 0, Data: map[int][]byte{}}
	component2 := game.PositionComponent{Position: math.NewVectorInt(5, 5)}
	component2.WriteUDP(&data2)

	packet := server.WorldState{}
	packet.Created = append(packet.Created, &data)
	packet.Updates = append(packet.Updates, &data2)

	world.Update(0.016)
	world.Update(0.016)
	world.Update(0.016)
	world.Update(0.016)
	world.Update(0.016)

	client.HandleWorldStatePacket(&packet, world,storage)

	assert.Equal(t, 1, len(world.Entities))

	position := world.Entities[0].Components[int(ecs.PositionComponentType)].(*game.PositionComponent)

	assert.Equal(t, 5, position.Position.X())
	assert.Equal(t, 5, position.Position.Y())

}

func TestRunClientUpdateAndCreateWithCacheAndInput(t *testing.T) {

	world, client, storage := createWorld()

	// create
	data := server.NetworkData{OwnerId: 0, NetworkId: 0, PrefabId: 0, Data: map[int][]byte{}}
	component := game.PositionComponent{Position: math.NewVectorInt(0, 0)}
	component.WriteUDP(&data)

	packet := server.WorldState{Tick: 0}
	packet.Created = append(packet.Created, &data)

	world.Update(0.016)

	client.HandleWorldStatePacket(&packet, world, storage)

	ecs.AddComponentsToStorage(world.Entities[0], map[int]*ecs.Storage{int(ecs.NetworkInstanceComponentType):storage})

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

	packet2 := server.WorldState{Tick: 1}
	packet2.Updates = append(packet.Updates, &data2)

	client.HandleWorldStatePacket(&packet2, world, storage)

	position1 := world.Entities[0].Components[int(ecs.PositionComponentType)].(*game.PositionComponent)
	assert.Equal(t, 0, position1.Position.X())
	assert.Equal(t, 0, position1.Position.Y())

}

func TestRunClientFromRealTest(t *testing.T) {

	world, client, storage := createWorld()

	packet := server.WorldState{Tick: 1244}

	world.SetToTick(1165)

	for i := 0; i < 83; i++ {
		world.Update(0.016)
	}

	assert.Equal(t, int64(1248), world.CurrentTick)

	bytes, _ := base64.StdEncoding.DecodeString("GP+NAwEC/44AAQIBAVgBBAABAVkBBAAAAAP/jgA=")
	data := server.NetworkData{OwnerId: 0, NetworkId: 0, PrefabId: 0, Data: map[int][]byte{0: bytes}}

	packet.Created = append(packet.Created, &data)

	client.HandleWorldStatePacket(&packet, world, storage)

	assert.Equal(t, 1, len(world.Entities))
}


func TestRunClientFromRealTestBothCreateAndUpdate(t *testing.T) {

	world, client, storage := createWorld()

	packet := server.WorldState{Tick: 1244}

	world.SetToTick(1165)

	for i := 0; i < 83; i++ {
		world.Update(0.016)
	}

	assert.Equal(t, int64(1248), world.CurrentTick)

	bytes, _ := base64.StdEncoding.DecodeString("GP+NAwEC/44AAQIBAVgBBAABAVkBBAAAAAP/jgA=")
	data := server.NetworkData{OwnerId: 0, NetworkId: 0, PrefabId: 0, Data: map[int][]byte{0: bytes}}

	packet.Created = append(packet.Created, &data)

	bytes2, _ := base64.StdEncoding.DecodeString("GP+NAwEC/44AAQIBAVgBBAABAVkBBAAAAAP/jgA=")
	data2 := server.NetworkData{OwnerId: 0, NetworkId: 0, PrefabId: 0, Data: map[int][]byte{0: bytes2}}

	packet.Updates = append(packet.Updates, &data2)

	client.HandleWorldStatePacket(&packet, world, storage)

	assert.Equal(t, 1, len(world.Entities))
}

func TestRunClientFromRealTestMoreComplete(t *testing.T) {

	worldClient, handlerClient, clientStorage := createWorld()
	worldClient.AddSystem(new(testNetworkSystem))

	handlerClient.PlayerId = 0

	worldServer, _, _ := createWorld()
	worldServer.AddSystem(new(server.NetworkInputFutureCollectionSystem))

	for i := 0; i < 32; i++ {
		worldServer.Update(0.016)
	}

	entity, e := worldServer.PrefabData.CreatePrefab(0)
	assert.Nil(t, e)
	netInstance := new(server.NetworkInstanceComponent)
	netInstance.NetworkId = 0
	netInstance.OwnerId = 0
	entity.Components[int(ecs.NetworkInstanceComponentType)] = netInstance

	worldServer.AddEntityToWorld(entity)
	worldServer.Input.Player[0] = ecs.NewInput()

	handlerClient.HandleHandshake(server.ServerConnectionHandshakePacket{PlayerId:0, RountTripClientTime:worldServer.CurrentTick}, worldClient)

	for i := 0; i < 3; i++ {
		worldServer.Update(0.016)
		worldClient.Update(0.016)
	}

	//create world state packet
	data := server.NetworkData{OwnerId: 0, NetworkId: 0, Data: map[int][]byte{}}
	component := game.PositionComponent{Position: math.NewVectorInt(2, 2)}
	component.WriteUDP(&data)

	// update
	data2 := server.NetworkData{OwnerId: 0, NetworkId: 0, Data: map[int][]byte{}}
	component2 := game.PositionComponent{Position: math.NewVectorInt(5, 5)}
	component2.WriteUDP(&data2)

	packet := server.WorldState{}
	packet.Created = append(packet.Created, &data)
	packet.Updates = append(packet.Updates, &data2)

	handlerClient.HandleWorldStatePacket(&packet, worldClient, clientStorage)

	worldClient.Input.Player[0].KeyPressed[ecs.Down] = true
	tickInputDown := worldClient.CurrentTick
	bytesInputDown := worldClient.Input.Player[0].ToBytes()
	worldClient.Update(0.016)
	worldServer.Update(0.016)

	worldClient.Update(0.016)
	worldServer.Update(0.016)

	worldServer.SetFutureInput(tickInputDown, bytesInputDown[0], 0)

	worldClient.Update(0.016)
	worldServer.Update(0.016)
	worldServer.SetFutureInput(tickInputDown+1, bytesInputDown[0], 0)
}
