package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteBytesUp(t *testing.T) {
	netInput := NetworkInput{
		Up:true,
		Down:false,
		Left:false,
		Right:false,
		X:false,
		C:false,
	}

	assert.Equal(t, []byte{0x1}, netInput.ToBytes());
}

func TestWriteBytesDown(t *testing.T) {
	netInput := NetworkInput{
		Up:false,
		Down:true,
		Left:false,
		Right:false,
		X:false,
		C:false,
	}

	assert.Equal(t, []byte{0x2}, netInput.ToBytes());
}

func TestWriteBytesLeft(t *testing.T) {
	netInput := NetworkInput{
		Up:false,
		Down:false,
		Right:false,
		Left:true,
		X:false,
		C:false,
	}

	assert.Equal(t, []byte{0x8}, netInput.ToBytes());
}

func TestWriteBytesRight(t *testing.T) {
	netInput := NetworkInput{
		Up:false,
		Down:false,
		Right:true,
		Left:false,
		X:false,
		C:false,
	}

	assert.Equal(t, []byte{0x4}, netInput.ToBytes());
}

func TestWriteBytesX(t *testing.T) {
	netInput := NetworkInput{
		Up:false,
		Down:false,
		Left:false,
		Right:false,
		X:true,
		C:false,
	}

	assert.Equal(t, []byte{0x10}, netInput.ToBytes());
}

func TestWriteBytesC(t *testing.T) {
	netInput := NetworkInput{
		Up:false,
		Down:false,
		Left:false,
		Right:false,
		X:false,
		C:true,
	}

	assert.Equal(t, []byte{0x20}, netInput.ToBytes());
}

func TestReadBytesUp(t *testing.T) {
	assert.True(t, NetworkInputFromBytes(0x1).Up);
}

func TestReadBytesDown(t *testing.T) {
	assert.True(t, NetworkInputFromBytes(0x2).Down);
}

func TestReadBytesLeft(t *testing.T) {
	assert.True(t, NetworkInputFromBytes(0x8).Left);
}

func TestReadBytesRight(t *testing.T) {
	assert.True(t, NetworkInputFromBytes(0x4).Right);
}

func TestReadBytesX(t *testing.T) {
	assert.True(t, NetworkInputFromBytes(0x10).X);
}

func TestReadBytesC(t *testing.T) {
	assert.True(t, NetworkInputFromBytes(0x20).C);
}

func TestReadMultiBytes(t *testing.T) {
	networkInput := NetworkInputFromBytes(0x15)
	assert.True(t, networkInput.Up);
	assert.True(t, networkInput.X);
	assert.True(t, networkInput.Right);
}
