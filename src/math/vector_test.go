package math

import (
	json2 "encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVectorIntParse(t *testing.T) {
	json := `{"Vec":[0,1]}`

	var test struct{
		Vec VectorInt
	}

	err := json2.Unmarshal([]byte(json), &test)

	assert.NoError(t, err)

	assert.Equal(t, 0,test.Vec.X())
	assert.Equal(t, 1, test.Vec.Y())
}

func TestVectorParse(t *testing.T) {
	json := `{"Vec":[0.0,1.0]}`

	var test struct{
		Vec Vector
	}

	err := json2.Unmarshal([]byte(json), &test)

	assert.NoError(t, err)

	assert.Equal(t, float64(0),test.Vec.X())
	assert.Equal(t, float64(1), test.Vec.Y())
}

func TestVectorInt_Set(t *testing.T) {
	vector := VectorInt{}

	vector.Set(1,1)

	assert.Equal(t, 1, vector.X())
	assert.Equal(t, 1, vector.Y())
}

func TestVector_Set(t *testing.T) {
	vector := Vector{}

	vector.Set(1.2,1.2)

	assert.Equal(t, float64(1.2), vector.X())
	assert.Equal(t, float64(1.2), vector.Y())
}

func TestVectorInt_Scale(t *testing.T) {
	vector := NewVectorInt(2,2)

	vector = vector.Scale(2)

	assert.Equal(t, 4, vector.X())
	assert.Equal(t, 4, vector.Y())
}

func TestVector_Scale(t *testing.T) {
	vector := NewVector(2.2,2.2)

	vector = vector.Scale(2.2)

	assert.True(t, float64(4.84) - vector.X() < 0.0001)
	assert.True(t, float64(4.84) - vector.Y() < 0.0001)
}

func TestVectorInt_Mul(t *testing.T) {
	vector := NewVectorInt(2,2)
	vector2 := NewVectorInt(2,3)

	vector = vector.Mul(vector2)

	assert.Equal(t, 4, vector.X())
	assert.Equal(t, 6, vector.Y())
}

func TestVectorInt_ClampHigher(t *testing.T) {
	vector := NewVector(2.2,2.66)
	vector2 := NewVector(1.1,1.1)

	vector = vector.Clamp(vector2, VectorZero())

	assert.True(t, float64(1.1) - vector.X() < 0.0001)
	assert.True(t, float64(1.1) - vector.Y() < 0.0001)
}

func TestVectorInt_ClampLower(t *testing.T) {
	vector := NewVector(2.2,2.66)
	vector2 := NewVector(5.1,5.1)

	vector = vector.Clamp(vector2, VectorZero())

	assert.True(t, float64(2.2) - vector.X() < 0.0001)
	assert.True(t, float64(2.2) - vector.Y() < 0.0001)
}

func TestVectorInt_ClampOneHigher(t *testing.T) {
	vector := NewVector(2.2,2.66)
	vector2 := NewVector(4.1,1.1)

	vector = vector.Clamp(vector2, VectorZero())

	assert.True(t, float64(2.2) - vector.X() < 0.0001)
	assert.True(t, float64(1.1) - vector.Y() < 0.0001)
}

func TestVectorInt_Remainder(t *testing.T) {
	vector := NewVector(2.2,2.66)

	vector = vector.Remaining()

	assert.True(t, float64(0.2) - vector.X() < 0.0001)
	assert.True(t, float64(0.66) - vector.Y() < 0.0001)
}

//func (self Vector) Remaining() Vector {
//
//
//func (self Vector) Round() Vector {
//
//}