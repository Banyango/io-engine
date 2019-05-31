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
	inp.KeyUp = raw.KeyUp
	inp.KeyPressed = raw.KeyPressed
	inp.KeyDown = raw.KeyDown
	inp.MousePosition = raw.MousePosition
	inp.MouseDown = raw.MouseDown
	inp.MouseUp = raw.MouseUp
	inp.MousePressed = raw.MousePressed
}

type RawInputGlobal struct {

	KeyDown    map[server.KeyCode]bool
	KeyPressed map[server.KeyCode]bool
	KeyUp      map[server.KeyCode]bool

	MousePosition math.Vector

	MouseDown    map[int]bool
	MousePressed map[int]bool
	MouseUp      map[int]bool

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

	self.KeyUp = map[server.KeyCode]bool{}
	self.KeyDown = map[server.KeyCode]bool{}
	self.KeyPressed = map[server.KeyCode]bool{}
	self.MouseDown = map[int]bool{}
	self.MousePressed = map[int]bool{}
	self.MouseUp = map[int]bool{}
	self.MousePosition = math.VectorZero()

	go func() {
		doc := js.Global().Get("document")

		if doc.Truthy() {
			done := make(chan struct{}, 0)

			self.mouseMoveFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				e := args[0]
				self.MousePosition.Set(e.Get("clientX").Float(), e.Get("clientY").Float())
				return nil;
			})
			defer self.mouseMoveFunc.Release()

			self.keyDownFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				e := args[0]

				//fmt.Println("key down:", e.Get("keyCode"))

				keyCode, err := KeyFromString(e.Get("keyCode").String())

				if err == nil {
					self.KeyDown[keyCode] = true
					self.KeyPressed[keyCode] = true
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
					self.KeyPressed[keyCode] = false
					self.KeyUp[keyCode] = true
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
	self.MousePressed = nil
	self.MouseDown = nil
	self.KeyDown = map[server.KeyCode]bool{}
	self.KeyPressed = map[server.KeyCode]bool{}
	self.KeyUp = map[server.KeyCode]bool{}
}

func (self *RawInputGlobal) AnyKeyPressed() bool {
	for _, i := range self.KeyPressed {
		if i {
			return true
		}
	}

	return false
}

func (self *RawInputGlobal) ToNetworkInput() *server.NetworkInput {
	return &server.NetworkInput{
		Up:self.KeyPressed[server.Up],
		Down:self.KeyPressed[server.Down],
		Right:self.KeyPressed[server.Right],
		Left:self.KeyPressed[server.Left],
		X:self.KeyPressed[server.X],
		C:self.KeyPressed[server.C],
	}
}

func KeyFromString(s string) (server.KeyCode, error) {
	keycode, err := strconv.Atoi(s)

	if err != nil {
		return -1, err
	}

	return server.KeyCode(keycode), err
}



