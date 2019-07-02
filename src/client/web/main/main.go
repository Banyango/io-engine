package main

import (
	"fmt"
	"github.com/Banyango/io-engine/src/client"
	"github.com/Banyango/io-engine/src/client/web"
	"github.com/Banyango/io-engine/src/game"
	"github.com/Banyango/io-engine/src/ecs"
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

	w.Log = new(web.WebLogger)

	input := new(web.ClientInputSystem)
	collision := new(game.CollisionSystem)
	movement  := new(client.ClientMovementSystem)
	netClient := new(web.NetworkedClientSystem)
	debugClient := new(web.DebugSystem)

	renderer := new(web.CanvasRenderSystem)

	w.AddRenderer(renderer)

	w.AddSystem(input)
	w.AddSystem(movement)
	w.AddSystem(collision)
	w.AddSystem(netClient)
	w.AddSystem(debugClient)

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
	var setTimeout js.Func

	self.CurrentFrameTime = time.Now().UnixNano() / int64(time.Millisecond)
	self.LastFrameTime = self.CurrentFrameTime
	self.TimeElapsed = 0

	defer func() {
		if r := recover(); r != nil {
			js.Global().Get("console").Call("log", "Recovering ", r)
		}
	}()

	setTimeout = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if !self.Paused {
			self.CurrentFrameTime = time.Now().UnixNano() / int64(time.Millisecond)
			delta := self.CurrentFrameTime - self.LastFrameTime
			self.TimeElapsed = self.TimeElapsed + delta
			for self.TimeElapsed >= self.Interval {
				self.Update(ecs.FIXED_DELTA)
				self.TimeElapsed = self.TimeElapsed - self.Interval
			}
			self.LastFrameTime = self.CurrentFrameTime
		}

		js.Global().Call("setTimeout", setTimeout, ecs.FIXED_DELTA)

		return nil
	})

	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {

		self.Render()

		js.Global().Call("requestAnimationFrame", renderFrame, 0.016)

		return nil
	})

	defer renderFrame.Release()

	js.Global().Call("requestAnimationFrame", renderFrame)
	js.Global().Call("setTimeout", setTimeout)

	<-done
}
