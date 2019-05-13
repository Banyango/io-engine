package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io-engine-backend/src/game"
	"io-engine-backend/src/server"
	"io-engine-backend/src/shared"
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
	w := shared.NewWorld()

	input := new(game.InputSystem)
	collision := new(game.CollisionSystem)
	movement := new(game.KeyboardMovementSystem)
	networkServer := new(server.ConnectionHandlerSystem)

	w.AddSystem(input)
	w.AddSystem(collision)
	w.AddSystem(movement)
	w.AddSystem(networkServer)

	pm, err := shared.NewPrefabManager(string(gameJson), w)

	if err != nil {
		log.Fatal(err)
	}

	w.PrefabData = pm

	w.CurrentFrameTime = time.Now().UnixNano() / int64(time.Millisecond)
	w.TimeElapsed = 0

	go mainLoop(w)

	gameServer := Server{World: w}

	http.HandleFunc("/connect", gameServer.ws)
	http.Handle("/", http.FileServer(http.Dir(".")))

	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}


type Server struct {
	World *shared.World
}

func mainLoop(w *shared.World) {
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

	self.createPlayer(conn)

}

func (self *Server) createPlayer(conn *websocket.Conn) {

	entity, err := self.World.PrefabData.CreatePrefab(0)

	if err != nil {
		fmt.Println("Error Occurred Creating Player Prefab")
		return
	}

	networkConnectionComponent := new(server.NetworkConnectionComponent)

	entity.Components[int(shared.NetworkConnectionComponentType)] = networkConnectionComponent

	entity.Id = self.World.FetchAndIncrementId()
	self.World.AddEntityToWorld(entity)

	networkConnectionComponent.TcpConnection = conn
	networkConnectionComponent.PlayerId = uint16(entity.Id)

	fmt.Println("Client is connected.. ", entity.Id)
	global := self.World.Globals[shared.ServerGlobalType].(*server.ServerGlobal)

	networkConnectionComponent.Handshake(uint16(entity.Id), global.BufferedChanges)

	global.NetworkSpawn(uint16(entity.Id),0, 1, true)

	go networkConnectionComponent.ReadPump()
	go networkConnectionComponent.WritePump()

}
