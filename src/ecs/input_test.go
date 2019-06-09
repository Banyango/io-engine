package ecs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteBytesUp(t *testing.T) {
	netInput := Input{KeyPressed: map[KeyCode]bool{
		Up: 	true,
		Down:  false,
		Left:  false,
		Right: false,
		X:     false,
		C:     false,
	},
	}

	assert.Equal(t, []byte{0x1}, netInput.ToBytes());
}

func TestWriteBytesDown(t *testing.T) {
	netInput := Input{KeyPressed: map[KeyCode]bool{
		Up:false,
		Down:true,
		Left:false,
		Right:false,
		X:false,
		C:false,
	},
	}

	assert.Equal(t, []byte{0x2}, netInput.ToBytes());
}

func TestWriteBytesLeft(t *testing.T) {
	netInput := Input{KeyPressed: map[KeyCode]bool{
		Up:false,
		Down:false,
		Right:false,
		Left:true,
		X:false,
		C:false,
	},
	}

	assert.Equal(t, []byte{0x8}, netInput.ToBytes());
}

func TestWriteBytesRight(t *testing.T) {
	netInput := Input{KeyPressed: map[KeyCode]bool{
		Up:false,
		Down:false,
		Right:true,
		Left:false,
		X:false,
		C:false,
	},
	}

	assert.Equal(t, []byte{0x4}, netInput.ToBytes());
}

func TestWriteBytesX(t *testing.T) {
	netInput := Input{KeyPressed: map[KeyCode]bool{
		Up:false,
		Down:false,
		Left:false,
		Right:false,
		X:true,
		C:false,
	},
	}

	assert.Equal(t, []byte{0x10}, netInput.ToBytes());
}

func TestWriteBytesC(t *testing.T) {
	netInput := Input{KeyPressed: map[KeyCode]bool{
		Up:false,
		Down:false,
		Left:false,
		Right:false,
		X:false,
		C:true,
	},
	}

	assert.Equal(t, []byte{0x20}, netInput.ToBytes());
}

func TestReadBytesUp(t *testing.T) {
	assert.True(t, NewInputFromBytes(0x1).KeyPressed[Up]);
}

func TestReadBytesDown(t *testing.T) {
	assert.True(t, NewInputFromBytes(0x2).KeyPressed[Down]);
}

func TestReadBytesLeft(t *testing.T) {
	assert.True(t, NewInputFromBytes(0x8).KeyPressed[Left]);
}

func TestReadBytesRight(t *testing.T) {
	assert.True(t, NewInputFromBytes(0x4).KeyPressed[Right]);
}

func TestReadBytesX(t *testing.T) {
	assert.True(t, NewInputFromBytes(0x10).KeyPressed[X]);
}

func TestReadBytesC(t *testing.T) {
	assert.True(t, NewInputFromBytes(0x20).KeyPressed[C]);
}

func TestReadMultiBytes(t *testing.T) {
	networkInput := NewInputFromBytes(0x15)
	assert.True(t, networkInput.KeyPressed[Up]);
	assert.True(t, networkInput.KeyPressed[X]);
	assert.True(t, networkInput.KeyPressed[Right]);
}
