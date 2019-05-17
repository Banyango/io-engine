package main

import (
	"fmt"
	"io-engine-backend/src/client"
	"io-engine-backend/src/game"
	"io-engine-backend/src/shared"
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
	w := shared.NewWorld()

	rawInput := new(client.ClientInputSystem)
	input := new(game.InputSystem)
	collision := new(game.CollisionSystem)
	movement := new(game.KeyboardMovementSystem)
	netClient := new(client.NetworkedClientSystem)
	//clientData := new(client.ClientNetworkDataSystem)

	renderer := new(client.CanvasRenderSystem)

	w.AddSystem(rawInput)
	w.AddSystem(input)
	w.AddSystem(collision)
	w.AddSystem(movement)
	w.AddSystem(netClient)
	//w.AddSystem(clientData)

	w.AddRenderer(renderer)

	pm, err := shared.NewPrefabManager(string(gameJsonValue), w)

	if err != nil {
		fmt.Println("Error creating prefab manager")
		os.Exit(1)
	}

	w.PrefabData = pm

	MainLoopClient(w)
}

func MainLoopClient(self *shared.World) {

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
			self.Update(0.022)
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