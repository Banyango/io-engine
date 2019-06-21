package ecs

import (
	"github.com/Banyango/io-engine/src/math"
)

type PlayerId uint16

type InputController struct {
	Player map[PlayerId]*Input
}

type BufferedInput struct {
	Tick int64
	Bytes map[PlayerId]byte
}

func (self *InputController) Clone() InputController {
	ic := InputController{}

	ic.Player = map[PlayerId]*Input{}

	for k, v := range self.Player {
		ic.Player[k] = v.Clone()
	}

	return ic
}

type Input struct {
	KeyDown    map[KeyCode]bool
	KeyPressed map[KeyCode]bool
	KeyUp      map[KeyCode]bool

	MousePosition math.Vector

	MouseDown    map[int]bool
	MousePressed map[int]bool
	MouseUp      map[int]bool
}

func (self *Input) Clone() *Input {
	i := NewInput()

	for k, v := range self.KeyPressed {
		i.KeyPressed[k] = v
	}

	for k, v := range self.KeyDown {
		i.KeyDown[k] = v
	}

	for k, v := range self.KeyUp {
		i.KeyUp[k] = v
	}

	for k, v := range self.MousePressed {
		i.MousePressed[k] = v
	}

	for k, v := range self.MouseUp {
		i.MouseUp[k] = v
	}

	for k, v := range self.MouseDown {
		i.MouseDown[k] = v
	}

	i.MousePosition = self.MousePosition

	return i
}

func NewInput() *Input {
	input := new(Input)

	input.KeyUp = map[KeyCode]bool{}
	input.KeyDown = map[KeyCode]bool{}
	input.KeyPressed = map[KeyCode]bool{}
	input.MouseDown = map[int]bool{}
	input.MousePressed = map[int]bool{}
	input.MouseUp = map[int]bool{}
	input.MousePosition = math.VectorZero()

	return input
}

func (self *Input) AnyKeyPressed() bool {
	for _, i := range self.KeyPressed {
		if i {
			return true
		}
	}

	return false
}

// Simplified Input - only uses up/dwn/left/right/ X and C
func (self *Input) ToBytes() []byte {
	value := byte(0)

	if self.KeyPressed[Up] {
		setBit(&value, 0)
	}

	if self.KeyPressed[Down] {
		setBit(&value, 1)
	}

	if self.KeyPressed[Right] {
		setBit(&value, 2)
	}

	if self.KeyPressed[Left] {
		setBit(&value, 3)
	}

	if self.KeyPressed[X] {
		setBit(&value, 4)
	}

	if self.KeyPressed[C] {
		setBit(&value, 5)
	}

	return []byte{value}
}

func NewInputFromBytes(value byte) Input {
	input := Input{KeyPressed: map[KeyCode]bool{}}

	input.KeyPressed[Up] = hasBit(value, 0)
	input.KeyPressed[Down] = hasBit(value, 1)
	input.KeyPressed[Right] = hasBit(value, 2)
	input.KeyPressed[Left] = hasBit(value, 3)
	input.KeyPressed[X] = hasBit(value, 4)
	input.KeyPressed[C] = hasBit(value, 5)

	return input
}

func (self *Input) InputFromBytes(value byte) {
	self.KeyPressed[Up] = hasBit(value, 0)
	self.KeyPressed[Down] = hasBit(value, 1)
	self.KeyPressed[Right] = hasBit(value, 2)
	self.KeyPressed[Left] = hasBit(value, 3)
	self.KeyPressed[X] = hasBit(value, 4)
	self.KeyPressed[C] = hasBit(value, 5)
}

func hasBit(n byte, pos uint) bool {
	val := n & (1 << pos)
	return val > 0
}

func setBit(n *byte, pos uint) {
	*n |= 1 << pos
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
