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

func TestSlerpInt(t *testing.T) {
	slerpInt := SlerpInt(NewVectorInt(0, 0), NewVectorInt(2, 2), 0.5)

	assert.Equal(t, 1, slerpInt.X())
	assert.Equal(t, 1, slerpInt.Y())
}

func TestSlerpIntNeg(t *testing.T) {
	slerpInt := SlerpInt(NewVectorInt(-2, -2), NewVectorInt(2, 2), 0.5)

	assert.Equal(t, 0, slerpInt.X())
	assert.Equal(t, 0, slerpInt.Y())

	slerpInt2 := SlerpInt(NewVectorInt(0, 0), NewVectorInt(-2, -2), 0.5)

	assert.Equal(t, -1, slerpInt2.X())
	assert.Equal(t, -1, slerpInt2.Y())

}

func TestWithinInt(t *testing.T) {
	assert.True(t, NewVectorInt(2,2).Within(NewVectorInt(3,3), 1))
	assert.False(t, NewVectorInt(2,2).Within(NewVectorInt(4,4), 1))
}

//func (self Vector) Remaining() Vector {
//
//
//func (self Vector) Round() Vector {
//
//}