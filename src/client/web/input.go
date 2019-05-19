package web

import (
	"github.com/goburrow/dynamic"
	"io-engine-backend/src/math"
	"io-engine-backend/src/server"
	. "io-engine-backend/src/ecs"
	"strconv"
	"syscall/js"
)

type ClientInputSystem struct {

}

func (self *ClientInputSystem) Init() {
	dynamic.Register("RawInputGlobal", func() interface{} {
		return &RawInputGlobal{}
	})
}

func (self *ClientInputSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{RawInputGlobalType}
}

func (self *ClientInputSystem) AddToStorage(entity *Entity) {

}

func (self *ClientInputSystem) UpdateSystem(delta float64, world *World) {

}

func Copy(inp *server.NetworkInputComponent, raw RawInputGlobal) {
	inp.KeyUp = raw.keyUp
	inp.KeyPressed = raw.keyPressed
	inp.KeyDown = raw.keyDown
	inp.MousePosition = raw.mousePosition
	inp.MouseDown = raw.mouseDown
	inp.MouseUp = raw.mouseUp
	inp.MousePressed = raw.mousePressed
}

type RawInputGlobal struct {

	keyDown map[server.KeyCode]bool
	keyPressed map[server.KeyCode]bool
	keyUp map[server.KeyCode]bool

	mousePosition math.Vector

	mouseDown map[int]bool
	mousePressed map[int]bool
	mouseUp map[int]bool

	keyDownFunc js.Func
	keyPressedFunc js.Func
	keyUpFunc js.Func

	mouseMoveFunc js.Func
	mouseDownFunc js.Func
	mouseUpFunc js.Func
}

func (self *RawInputGlobal) Id() int {
	return int(RawInputGlobalType)
}

func (self *RawInputGlobal) CreateGlobal(world *World) {

	self.keyUp = map[server.KeyCode]bool{}
	self.keyDown = map[server.KeyCode]bool{}
	self.keyPressed = map[server.KeyCode]bool{}
	self.mouseDown = map[int]bool{}
	self.mousePressed = map[int]bool{}
	self.mouseUp = map[int]bool{}
	self.mousePosition = math.VectorZero()

	go func() {
		doc := js.Global().Get("document")

		if doc.Truthy() {
			done := make(chan struct{}, 0)

			self.mouseMoveFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				e := args[0]
				self.mousePosition.Set(e.Get("clientX").Float(), e.Get("clientY").Float())
				return nil;
			})
			defer self.mouseMoveFunc.Release()

			self.keyDownFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				e := args[0]

				//fmt.Println("key down:", e.Get("keyCode"))

				keyCode, err := KeyFromString(e.Get("keyCode").String())

				if err == nil {
					self.keyDown[keyCode] = true
					self.keyPressed[keyCode] = true
				}

				return nil;
			})

			defer self.keyDownFunc.Release()

			defer self.keyPressedFunc.Release()

			self.keyUpFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				e := args[0]

				//fmt.Println("key up:", e.Get("keyCode"))

				keyCode, err := KeyFromString(e.Get("keyCode").String())

				if err == nil {
					self.keyPressed[keyCode] = false
					self.keyUp[keyCode] = true
				}

				return nil;
			})

			defer self.keyUpFunc.Release()

			doc.Call("addEventListener", "mousemove", self.mouseMoveFunc)
			doc.Call("addEventListener", "keydown", self.keyDownFunc)
			doc.Call("addEventListener", "keyup", self.keyUpFunc)

			<-done
		}
	}()
}

func (self *RawInputGlobal) Reset() {
	self.mousePressed = nil
	self.mouseDown = nil
	self.keyDown = map[server.KeyCode]bool{}
	self.keyPressed = map[server.KeyCode]bool{}
	self.keyUp = map[server.KeyCode]bool{}
}

func (self *RawInputGlobal) ToNetworkInput() *server.NetworkInput {
	return &server.NetworkInput{
		Up:self.keyPressed[server.Up],
		Down:self.keyPressed[server.Down],
		Right:self.keyPressed[server.Right],
		Left:self.keyPressed[server.Left],
		X:self.keyPressed[server.X],
		C:self.keyPressed[server.C],
	}
}

func KeyFromString(s string) (server.KeyCode, error) {
	keycode, err := strconv.Atoi(s)

	if err != nil {
		return -1, err
	}

	return server.KeyCode(keycode), err
}



