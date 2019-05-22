package main

import (
	"fmt"
	"github.com/Banyango/socker"
	"github.com/gorilla/websocket"
	"io-engine-backend/src/ecs"
	"io-engine-backend/src/game"
	"io-engine-backend/src/server"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	gameJson, err := ioutil.ReadFile("./game.json");

	if err != nil {
		os.Exit(1)
	}

	fmt.Println("Creating World.")
	w := ecs.NewWorld()

	input := new(server.InputSystem)
	collision := new(game.CollisionSystem)
	movement := new(game.KeyboardMovementSystem)
	playerState := new(server.PlayerStateSystem)
	networkServer := new(server.ConnectionHandlerSystem)
	networkInstance := new(server.NetworkInstanceDataCollectionSystem)

	w.AddSystem(networkInstance)
	w.AddSystem(networkServer)
	w.AddSystem(playerState)
	w.AddSystem(input)
	w.AddSystem(movement)
	w.AddSystem(collision)

	pm, err := ecs.NewPrefabManager(string(gameJson), w)

	if err != nil {
		log.Fatal(err)
	}

	w.PrefabData = pm

	w.CurrentFrameTime = time.Now().UnixNano() / int64(time.Millisecond)
	w.TimeElapsed = 0

	go mainLoop(w)

	gameServer := Server{World: w}

	http.HandleFunc("/connect", gameServer.ws)
	http.Handle("/", http.FileServer(http.Dir("./src/client/web/main/")))
	http.HandleFunc("/game.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./game.json")
	})

	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}

type Server struct {
	World *ecs.World
}

func mainLoop(w *ecs.World) {
	for true {

		w.LastFrameTime = w.CurrentFrameTime

		w.CurrentFrameTime = time.Now().UnixNano() / int64(time.Millisecond)

		delta := w.CurrentFrameTime - w.LastFrameTime

		w.TimeElapsed = w.TimeElapsed + delta

		for w.TimeElapsed >= w.Interval {
			w.Update(0.016)
			w.TimeElapsed = w.TimeElapsed - w.Interval
		}
	}
}

func (self *Server) ws(writer http.ResponseWriter, request *http.Request)  {
	upgrader := websocket.Upgrader{}

	fmt.Println("Client is connecting..")

	conn, err := upgrader.Upgrade(writer, request, nil)

	if err != nil {
		fmt.Println("Error Occurred")
		return
	}

	self.createClientConnection(conn)

}

func (self *Server) createClientConnection(conn *websocket.Conn) {

	entity, err := self.World.PrefabData.CreatePrefab(0)

	if err != nil {
		fmt.Println("Error Occurred Creating Player Prefab")
		return
	}

	networkConnectionComponent := new(server.NetworkConnectionComponent)
	entity.Components[int(ecs.NetworkConnectionComponentType)] = networkConnectionComponent

	networkInputComponent := new(server.NetworkInputComponent)
	entity.Components[int(ecs.NetworkInputComponentType)] = networkInputComponent


	entity.Id = self.World.FetchAndIncrementId()
	self.World.AddEntityToWorld(entity)

	global := self.World.Globals[ecs.ServerGlobalType].(*server.ServerGlobal)

	networkConnectionComponent.WSConnHandler = socker.NewClientConnection(conn)
	networkConnectionComponent.PlayerId = global.FetchAndIncrementPlayerId()

	inputGlobal := self.World.Globals[int(ecs.NetworkInputGlobalType)].(*server.NetworkInputGlobal)
	inputGlobal.Inputs[server.NetworkId(networkConnectionComponent.PlayerId)] = networkInputComponent

	fmt.Println("Client entityId: ", entity.Id, " Given playerId: ", networkConnectionComponent.PlayerId)

	networkConnectionComponent.WSConnHandler.Add(func(message []byte) bool {
		fmt.Println("Sending Buffered Entities to -> player ", networkConnectionComponent.PlayerId)
		networkConnectionComponent.SendBufferedEntites(global.BufferedEntityChanges)
		return true
	})

	// setup webrtc connection send offer
	networkConnectionComponent.WSConnHandler.Add(func(message []byte) bool {
		fmt.Println("Setting up webrtc data channel -> player ", networkConnectionComponent.PlayerId)
		networkConnectionComponent.ConnectToDataChannel(message)
		return true
	})

	networkConnectionComponent.WSConnHandler.Add(func(message []byte) bool {
		fmt.Println("Handling messages -> player", networkConnectionComponent.PlayerId)
		networkConnectionComponent.HandleSignal(message)
		return networkConnectionComponent.IsDataChannelOpen
	})

	networkConnectionComponent.WSConnHandler.Add(func(message []byte) bool {
		networkConnectionComponent.GameMessage(message)
		return false
	})

	go networkConnectionComponent.WSConnHandler.ReadPump()
	go networkConnectionComponent.WSConnHandler.WritePump()

}

