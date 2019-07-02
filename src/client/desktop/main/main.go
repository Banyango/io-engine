package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/Banyango/io-engine/src/client/web"
	"github.com/Banyango/io-engine/src/ecs"
	"github.com/Banyango/io-engine/src/game"
	"io/ioutil"
	"os"
	"time"
)

func main() {

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	defer sdl.Quit()

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	defer window.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}

	surface.FillRect(nil, 0)

	rect := sdl.Rect{0, 0, 200, 200}
	surface.FillRect(&rect, 0xffff0000)
	window.UpdateSurface()

	gameJson, err := ioutil.ReadFile("../../../../game.json");

	fmt.Println("Creating World.")
	w := ecs.NewWorld()

	rawInput := new(web.ClientInputSystem)
	collision := new(game.CollisionSystem)
	netClient := new(web.NetworkedClientSystem)
	//clientData := new(client.ClientNetworkDataSystem)

	renderer := new(web.CanvasRenderSystem)

	w.AddSystem(rawInput)
	w.AddSystem(collision)
	w.AddSystem(netClient)
	//w.AddSystem(clientData)

	w.AddRenderer(renderer)

	pm, err := ecs.NewPrefabManager(string(gameJson), w)

	if err != nil {
		fmt.Println("Error creating prefab manager")
		os.Exit(1)
	}

	w.PrefabData = pm

	w.CurrentFrameTime = time.Now().UnixNano() / int64(time.Millisecond)
	w.TimeElapsed = 0

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
	}
}

func RenderFrame(self *ecs.World) {
	self.LastFrameTime = self.CurrentFrameTime

	self.CurrentFrameTime = time.Now().UnixNano() / int64(time.Millisecond)

	delta := self.CurrentFrameTime - self.LastFrameTime

	self.TimeElapsed = self.TimeElapsed + delta

	for self.TimeElapsed >= self.Interval {
		self.Update(0.022)
		self.TimeElapsed = self.TimeElapsed - self.Interval
	}

	self.Render()
}
