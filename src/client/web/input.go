package web

import (
	. "io-engine-backend/src/ecs"
	"strconv"
	"syscall/js"
)

type ClientInputSystem struct {
	keyDownFunc js.Func
	keyPressedFunc js.Func
	keyUpFunc js.Func

	mouseMoveFunc js.Func
	mouseDownFunc js.Func
	mouseUpFunc js.Func

	callbackInput *Input
}

func (self *ClientInputSystem) Init(world *World) {
	go func() {
		doc := js.Global().Get("document")

		if doc.Truthy() {
			done := make(chan struct{}, 0)

			self.mouseMoveFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				e := args[0]
				self.callbackInput.MousePosition.Set(e.Get("clientX").Float(), e.Get("clientY").Float())
				return nil;
			})
			defer self.mouseMoveFunc.Release()

			self.keyDownFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				e := args[0]

				//fmt.Println("key down:", e.Get("keyCode"))

				keyCode, err := KeyFromString(e.Get("keyCode").String())

				if err == nil {
					self.callbackInput.KeyDown[keyCode] = true
					self.callbackInput.KeyPressed[keyCode] = true
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
					self.callbackInput.KeyPressed[keyCode] = false
					self.callbackInput.KeyUp[keyCode] = true
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

func (self *ClientInputSystem) RemoveFromStorage(entity *Entity) {

}

func (self *ClientInputSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{}
}

func (self *ClientInputSystem) AddToStorage(entity *Entity) {

}

func (self *ClientInputSystem) UpdateSystem(delta float64, world *World) {

	if world.IsResimulating {
		return
	}

	clone := self.callbackInput.Clone()

	world.Input.Player[0] = &clone

	self.callbackInput = NewInput()

}

func KeyFromString(s string) (KeyCode, error) {
	keycode, err := strconv.Atoi(s)

	if err != nil {
		return -1, err
	}

	return KeyCode(keycode), err
}



