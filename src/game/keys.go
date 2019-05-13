package game

import (
	"github.com/goburrow/dynamic"
	"io-engine-backend/src/math"
	"io-engine-backend/src/shared"
	"strconv"
)

type InputSystem struct {

}

func (*InputSystem) Init() {
	dynamic.Register("InputGlobal", func() interface{} {
		return &InputGlobal{}
	})
}

func (*InputSystem) AddToStorage(entity shared.Entity) {

}

func (*InputSystem) RequiredComponentTypes() []shared.ComponentType {
	return []shared.ComponentType{}
}

func (*InputSystem) UpdateSystem(delta float64, world *shared.World) {

}

type InputGlobal struct {
	KeyDown    map[KeyCode]bool
	KeyPressed map[KeyCode]bool
	KeyUp      map[KeyCode]bool

	MousePosition math.Vector

	MouseDown    map[int]bool
	MousePressed map[int]bool
	MouseUp      map[int]bool
}

func (self *InputGlobal) Id() int {
	return int(shared.InputGlobalType)
}

func (self *InputGlobal) CreateGlobal(world *shared.World) {
	self.KeyUp = map[KeyCode]bool{}
	self.KeyDown = map[KeyCode]bool{}
	self.KeyPressed = map[KeyCode]bool{}
	self.MouseDown = map[int]bool{}
	self.MousePressed = map[int]bool{}
	self.MouseUp = map[int]bool{}
	self.MousePosition = math.VectorZero()
}

func (self *InputGlobal) Any() bool {
	return false
}

func (self *InputGlobal) AnyKeyPressed() bool {
	for _, i := range self.KeyPressed {
		if i {
			return true
		}
	}

	return false
}

type KeyCode int

func KeyFromString(s string) (KeyCode, error) {
	keycode, err := strconv.Atoi(s)

	if err != nil {
		return -1, err
	}

	return KeyCode(keycode), err
}
const (
	Up    KeyCode = 38
	Down  KeyCode = 40
	Right KeyCode = 37
	Left  KeyCode = 39
	X     KeyCode = 88
	C     KeyCode = 67
)
