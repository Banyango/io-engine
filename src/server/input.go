package server

import (
	"encoding/json"
	"fmt"
	"github.com/goburrow/dynamic"
	"io-engine-backend/src/ecs"
	"io-engine-backend/src/math"
)

type InputSystem struct {

}

func (*InputSystem) Init() {
	dynamic.Register("NetworkInputComponent", func() interface{} {
		return &NetworkInputComponent{}
	})
}

func (*InputSystem) AddToStorage(entity *ecs.Entity) {

}

func (*InputSystem) RequiredComponentTypes() []ecs.ComponentType {
	return []ecs.ComponentType{ecs.NetworkInputComponentType}
}

func (*InputSystem) UpdateSystem(delta float64, world *ecs.World) {

}

type NetworkInputComponent struct {
	KeyDown    map[KeyCode]bool
	KeyPressed map[KeyCode]bool
	KeyUp      map[KeyCode]bool

	MousePosition math.Vector

	MouseDown    map[int]bool
	MousePressed map[int]bool
	MouseUp      map[int]bool
}



func (self *NetworkInputComponent) Id() int {
	return int(ecs.NetworkInputComponentType)
}

func (self *NetworkInputComponent) CreateComponent() {
	self.KeyUp = map[KeyCode]bool{}
	self.KeyDown = map[KeyCode]bool{}
	self.KeyPressed = map[KeyCode]bool{}
	self.MouseDown = map[int]bool{}
	self.MousePressed = map[int]bool{}
	self.MouseUp = map[int]bool{}
	self.MousePosition = math.VectorZero()
}

func (self *NetworkInputComponent) DestroyComponent() {

}

func (self *NetworkInputComponent) Clone() ecs.Component {
	return new(NetworkInputComponent)
}

func (self *NetworkInputComponent) Any() bool {
	return false
}

func (self *NetworkInputComponent) AnyKeyPressed() bool {
	for _, i := range self.KeyPressed {
		if i {
			return true
		}
	}

	return false
}

func (self *NetworkInputComponent) HandleClientInput(bytes []byte) {

	fromBytes := NetworkInputFromBytes(bytes[0])

	self.SetKeyPresses(fromBytes, Up)
	self.SetKeyPresses(fromBytes, Down)
	self.SetKeyPresses(fromBytes, Left)
	self.SetKeyPresses(fromBytes, Right)
	self.SetKeyPresses(fromBytes, X)
	self.SetKeyPresses(fromBytes, C)

	marshal, _ := json.Marshal(self)
	fmt.Println(string(marshal))

}

func (self *NetworkInputComponent) SetKeyPresses(fromBytes NetworkInput, code KeyCode) {
	if !self.KeyPressed[code] && fromBytes.Up {
		self.KeyDown[code] = true
	}
	if self.KeyPressed[code] && !fromBytes.Up {
		self.KeyDown[code] = false
	}
	self.KeyPressed[code] = fromBytes.Up
}

type KeyCode int

const (
	Up    KeyCode = 38
	Down  KeyCode = 40
	Right KeyCode = 37
	Left  KeyCode = 39
	X     KeyCode = 88
	C     KeyCode = 67
)

