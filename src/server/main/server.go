package main

import (
	"fmt"
	"github.com/Banyango/io-engine/src/ecs"
	"github.com/Banyango/io-engine/src/game"
	"github.com/Banyango/io-engine/src/server"
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

	w.Log = new(server.ServerLogger)

	gameServer := server.Server{World: w}

	netInputBuffer := new(server.NetworkInputFutureCollectionSystem)
	movement := new(game.KeyboardMovementSystem)
	collision := new(game.CollisionSystem)
	spawn := new(game.SpawnSystem)
	networkCollect := server.NewNetworkInstanceDataCollectionSystem(&gameServer)

	w.AddSystem(netInputBuffer)
	w.AddSystem(movement)
	w.AddSystem(collision)
	w.AddSystem(spawn)
	w.AddSystem(networkCollect)

	pm, err := ecs.NewPrefabManager(string(gameJson), w)

	if err != nil {
		log.Fatal(err)
	}

	w.PrefabData = pm

	w.CurrentFrameTime = time.Now().UnixNano() / int64(time.Millisecond)
	w.TimeElapsed = 0

	spawn.AddSpawnListener(&gameServer)

	go mainLoop(&gameServer)

	http.HandleFunc("/connect", gameServer.Ws)
	http.Handle("/", http.FileServer(http.Dir("./src/client/web/main/")))
	http.HandleFunc("/game.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./game.json")
	})

	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}

func mainLoop(gameServer *server.Server) {
	for true {

		gameServer.World.LastFrameTime = gameServer.World.CurrentFrameTime

		gameServer.World.CurrentFrameTime = time.Now().UnixNano() / int64(time.Millisecond)

		delta := gameServer.World.CurrentFrameTime - gameServer.World.LastFrameTime

		gameServer.World.TimeElapsed = gameServer.World.TimeElapsed + delta

		for gameServer.World.TimeElapsed >= gameServer.World.Interval {
			gameServer.HandleIncomingData(0.016)
			gameServer.World.Update(0.016)
			gameServer.SendNetworkData(0.016)
			gameServer.World.TimeElapsed = gameServer.World.TimeElapsed - gameServer.World.Interval
		}
	}
}


