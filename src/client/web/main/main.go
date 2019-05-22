package main

import (
	"fmt"
	"io-engine-backend/src/client/web"
	"io-engine-backend/src/game"
	"io-engine-backend/src/ecs"
	"os"
	"syscall/js"
	"time"
)

func main() {

	gameJson := make(chan string)

	js.Global().Call("fetch", "/game.json").Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return args[0].Call("json")
	})).Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		js.Global().Get("console").Call("log", args[0])
		gameJson <- js.Global().Get("JSON").Call("stringify", args[0]).String()
		return nil
	}))

	gameJsonValue := <-gameJson

	fmt.Println("Creating World.")
	w := ecs.NewWorld()

	rawInput := new(web.ClientInputSystem)
	collision := new(game.CollisionSystem)
	netClient := new(web.NetworkedClientSystem)
	debugClient := new(web.DebugSystem)
	//clientData := new(client.ClientNetworkDataSystem)

	renderer := new(web.CanvasRenderSystem)

	w.AddSystem(rawInput)
	w.AddSystem(collision)
	w.AddSystem(netClient)
	w.AddSystem(debugClient)
	//w.AddSystem(clientData)

	w.AddRenderer(renderer)

	pm, err := ecs.NewPrefabManager(string(gameJsonValue), w)

	if err != nil {
		fmt.Println("Error creating prefab manager")
		os.Exit(1)
	}

	w.PrefabData = pm

	MainLoopClient(w)
}

func MainLoopClient(self *ecs.World) {

	done := make(chan struct{}, 0)

	var renderFrame js.Func

	self.CurrentFrameTime = time.Now().UnixNano() / int64(time.Millisecond)
	self.TimeElapsed = 0

	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {

		self.LastFrameTime = self.CurrentFrameTime

		self.CurrentFrameTime = time.Now().UnixNano() / int64(time.Millisecond)

		delta := self.CurrentFrameTime - self.LastFrameTime

		self.TimeElapsed = self.TimeElapsed + delta

		for self.TimeElapsed >= self.Interval {
			self.Update(0.016)
			self.TimeElapsed = self.TimeElapsed - self.Interval
		}

		self.Render()

		js.Global().Call("requestAnimationFrame", renderFrame)

		return nil
	})

	defer renderFrame.Release()

	js.Global().Call("requestAnimationFrame", renderFrame)

	<-done
}