package math

import (
	"encoding/json"
	"math"
)

/*

	Contains
		- Vector - a 2d float vector
		- VectorInt - a 2d int vector

 */


/*

	2D Vector

 */
type Vector struct {
	position [2]float64
}

func VectorZero() Vector {
	return Vector{position: [2]float64{0.0, 0.0}}
}

func VectorOne() Vector {
	return Vector{position: [2]float64{1.0, 1.0}}
}

func VectorUp() Vector {
	return Vector{position: [2]float64{0.0, -1.0}}
}

func VectorDown() Vector {
	return Vector{position: [2]float64{0.0, 1.0}}
}

func VectorLeft() Vector {
	return Vector{position: [2]float64{-1.0, 0.0}}
}

func VectorRight() Vector {
	return Vector{position: [2]float64{1.0, 0.0}}
}

func NewVector(x float64, y float64) Vector {
	return Vector{position: [2]float64{x, y}}
}

func (self *Vector) X() float64 {
	return self.position[0];
}

func (self *Vector) Y() float64 {
	return self.position[1];
}

func (self *Vector) UnmarshalJSON(b []byte) error {
	data := new([2]float64)

	err := json.Unmarshal(b, &data)

	if err != nil {
		return err
	}

	self.Set(data[0], data[1])

	return nil
}

func (self *Vector) Set(x float64, y float64) {
	self.position[0] = x
	self.position[1] = y
}

func (self Vector) Scale(value float64) Vector {
	for i := 0; i < 2; i++ {
		self.position[i] *= value
	}
	return self
}

func (self Vector) Add(value Vector) Vector {
	for i := 0; i < 2; i++ {
		self.position[i] += value.position[i]
	}
	return self
}

func (self Vector) Sub(value Vector) Vector {
	for i := 0; i < 2; i++ {
		self.position[i] -= value.position[i]
	}
	return self
}

func (self Vector) Mul(value Vector) Vector {
	for i := 0; i < 2; i++ {
		self.position[i] *= value.position[i]
	}
	return self
}

func (self Vector) Div(value Vector) Vector {
	for i := 0; i < 2; i++ {
		self.position[i] /= value.position[i]
	}
	return self
}

func (self Vector) Clamp(clampMin Vector, clampMax Vector) Vector {

	x := self.X()
	y := self.Y()

	if x > clampMax.X() {
		x = clampMax.X()
	}

	if y > clampMax.Y() {
		y = clampMax.Y()
	}

	if x < clampMin.X() {
		x = clampMin.X()
	}

	if y < clampMin.Y() {
		y = clampMin.Y()
	}

	return NewVector(x,y)
}

func (self Vector) Remaining() Vector {
	x := math.Trunc(self.position[0])
	y := math.Trunc(self.position[1])

	return self.Sub(NewVector(x,y))
}

func (self Vector) Truncate() Vector {
	x := math.Trunc(self.position[0])
	y := math.Trunc(self.position[1])

	return NewVector(x,y)
}

func (self Vector) Round() Vector {
	x := math.Round(self.position[0])
	y := math.Round(self.position[1])

	return NewVector(x,y)
}

func (self Vector) ToInt() VectorInt {
	return NewVectorInt(int(self.X()), int(self.Y()))
}

func (self Vector) Neg() Vector {
	return NewVector(-self.position[0], -self.position[1])
}








/*

	VectorInt

 */

type VectorInt struct {
	position [2]int
}

func VectorIntZero() VectorInt {
	return VectorInt{position: [2]int{0, 0}}
}

func VectorIntOne() VectorInt {
	return VectorInt{position: [2]int{1, 1}}
}

func VectorIntUp() VectorInt {
	return VectorInt{position: [2]int{0, -1}}
}

func VectorIntDown() VectorInt {
	return VectorInt{position: [2]int{0, 1}}
}

func NewVectorInt(x int, y int) VectorInt {
	return VectorInt{position: [2]int{x, y}}
}

func (self *VectorInt) X() int {
	return self.position[0];
}

func (self *VectorInt) Y() int {
	return self.position[1];
}

func (self *VectorInt) UnmarshalJSON(b []byte) error {
	data := new([2]int)

	err := json.Unmarshal(b, &data)

	if err != nil {
		return err
	}

	self.Set(data[0], data[1])

	return nil
}

//Set mutates the vector values
func (self *VectorInt) Set(x int, y int) {
	self.position[0] = x
	self.position[1] = y
}

//Scale multiply by a scalar int
func (self VectorInt) Scale(value int) VectorInt {
	for i := 0; i < 2; i++ {
		self.position[i] *= value
	}
	return self
}

func (self VectorInt) Add(value VectorInt) VectorInt {
	for i := 0; i < 2; i++ {
		self.position[i] += value.position[i]
	}
	return self
}

func (self VectorInt) Sub(value VectorInt) VectorInt {
	for i := 0; i < 2; i++ {
		self.position[i] -= value.position[i]
	}
	return self
}

func (self VectorInt) Mul(value VectorInt) VectorInt {
	for i := 0; i < 2; i++ {
		self.position[i] *= value.position[i]
	}
	return self
}

func (self VectorInt) Div(value VectorInt) VectorInt {
	for i := 0; i < 2; i++ {
		self.position[i] /= value.position[i]
	}
	return self
}
