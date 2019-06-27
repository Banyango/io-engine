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

func (self *Vector) MarshalJSON() ([]byte, error) {
	var data struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	data.X = self.position[0]
	data.Y = self.position[1]

	bytes, e := json.Marshal(data)

	if e != nil {
		return nil, e
	}

	return bytes, nil
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

	x := self.position[0]
	y := self.position[1]

	if x > clampMax.position[0] {
		x = clampMax.position[0]
	}

	if y > clampMax.position[1] {
		y = clampMax.position[1]
	}

	if x < clampMin.X() {
		x = clampMin.position[0]
	}

	if y < clampMin.position[1] {
		y = clampMin.position[1]
	}

	return NewVector(x, y)
}

func (self Vector) Remaining() Vector {
	x := math.Trunc(self.position[0])
	y := math.Trunc(self.position[1])

	return self.Sub(NewVector(x, y))
}

func (self Vector) Truncate() Vector {
	x := math.Trunc(self.position[0])
	y := math.Trunc(self.position[1])

	return NewVector(x, y)
}

func (self Vector) Round() Vector {
	x := math.Round(self.position[0])
	y := math.Round(self.position[1])

	return NewVector(x, y)
}

func (self Vector) ToInt() VectorInt {
	return NewVectorInt(int(self.position[0]), int(self.position[1]))
}

func (self Vector) Neg() Vector {
	return NewVector(-self.position[0], -self.position[1])
}

func (self Vector) Lerp(other Vector, time float64) Vector {
	return NewVector(Lerp(other.position[0], self.position[0], time), Lerp(other.position[1], self.position[1], time))
}

func (self Vector) Normalize() Vector {
	distance := math.Sqrt(self.position[0]*self.position[0] + self.position[1]*self.position[1])
	return NewVector(self.position[0]/distance, self.position[1]/distance)
}

func (self *Vector) Equals(vector Vector) bool {
	return math.Abs(float64(vector.position[1] - self.position[0])) < 0.001 &&
		math.Abs(float64(vector.position[1] - self.position[1])) < 0.001
}

func (self *Vector) Within(vector Vector, value float64) bool {
	return math.Abs(float64(vector.position[1] - self.position[0])) <= value &&
		math.Abs(float64(vector.position[1] - self.position[1])) <= value
}

func (self *Vector) ToVecInt() VectorInt {
	return NewVectorInt(int(math.Round(self.position[0])), int(math.Round(self.position[1])))
}

func Lerp(v0 float64, v1 float64, t float64) float64 {
	return v0*(1.0-t) + v1*t
}

func Dot(start Vector, end Vector) float64 {
	return start.position[0] * end.position[0] + start.position[1] * end.position[1]
}

func Slerp(start Vector, end Vector, percent float64) Vector {
	dot := Dot(start, end)

	Clamp(&dot, -1, 1)

	theta := math.Acos(dot) * percent

	relativeVec := end.Sub(start.Scale(dot))

	relativeVec.Normalize()

	return (start.Scale(math.Cos(theta))).Add(relativeVec.Scale(math.Sin(theta)))
}

func SlerpInt(start VectorInt, end VectorInt, percent float64) VectorInt {
	slerp := Slerp(start.ToVec(), end.ToVec(), percent)
	return slerp.ToInt()
}

func Clamp(v0 *float64, min float64, max float64) {
	if *v0 < min {
		*v0 = min
	}
	if *v0 > max {
		*v0 = max
	}
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

func (self *VectorInt) MarshalJSON() ([]byte, error) {
	var data struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	data.X = self.position[0]
	data.Y = self.position[1]

	bytes, e := json.Marshal(data)

	if e != nil {
		return nil, e
	}

	return bytes, nil
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

func (self VectorInt) ToVec() Vector {
	return NewVector(float64(self.position[0]), float64(self.position[1]))
}

func (self VectorInt) Within(other VectorInt, value int) bool {
	return math.Abs(float64(self.position[0] - other.position[0])) <= float64(value) &&
		math.Abs(float64(self.position[1] - other.position[1])) <= float64(value)
}
