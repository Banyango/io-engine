package web

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	. "github.com/Banyango/io-engine/src/ecs"
	"github.com/Banyango/io-engine/src/game"
	"github.com/Banyango/io-engine/src/math"
	"github.com/goburrow/dynamic"
	. "github.com/lucasb-eyer/go-colorful"
	math2 "math"
	"reflect"
	"syscall/js"
)

type CanvasRenderSystem struct {
	circleComponents   Storage
	positionComponents Storage
	lerpingComponents  map[int64]math.Vector

	CanvasElementId string
	Width           int
	Height          int
	Scale           math.Vector

	doc           js.Value
	canvasElement js.Value
	ctx           js.Value

	ShouldRender bool
}

func (self *CanvasRenderSystem) Init(w *World) {
	dynamic.Register("CircleRendererComponent", func() interface{} {
		return &CircleRendererComponent{}
	})

	self.circleComponents = NewStorage()
	self.positionComponents = NewStorage()
	self.lerpingComponents = map[int64]math.Vector{}

	self.Width = 600
	self.Height = 500
	self.CanvasElementId = "mycanvas"

	self.doc = js.Global().Get("document")

	if self.doc.Truthy() {
		fmt.Println("Setting Up Canvas Renderer on element: ", self.CanvasElementId)

		self.canvasElement = js.Global().Get("document").Call("getElementById", self.CanvasElementId)
		self.canvasElement.Set("width", self.Width)
		self.canvasElement.Set("height", self.Height)
		self.ctx = self.canvasElement.Call("getContext", "2d")
	}

}

func (self *CanvasRenderSystem) RemoveFromStorage(entity *Entity) {
	storages := map[int]*Storage{
		int(PositionComponentType): &self.positionComponents,
		int(CircleComponentType):   &self.circleComponents,
	}
	RemoveComponentsFromStorage(entity, storages)
	delete (self.lerpingComponents, entity.Id)
	self.ctx.Call("clearRect", 0, 0, self.Width, self.Height)
	self.ctx.Call("save")
}

func (self *CanvasRenderSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{PositionComponentType, CircleComponentType}
}

func (self *CanvasRenderSystem) UpdateFrequency() int {
	return 1
}

func (self *CanvasRenderSystem) AddToStorage(entity *Entity) {

	AddComponentsToStorage(entity, map[int]*Storage{
		int(PositionComponentType):&self.positionComponents,
		int(CircleComponentType):&self.circleComponents,
	})

	position := (*self.positionComponents.Components[entity.Id]).(*game.PositionComponent)
	self.lerpingComponents[entity.Id] = math.NewVectorInt(position.Position.X(), position.Position.Y()).ToVec()
}

func (self *CanvasRenderSystem) UpdateSystem(delta float64, world *World) {

	if world.IsResimulating {
		return
	}

	self.ctx.Call("clearRect", 0, 0, self.Width, self.Height)

	for entity, _ := range self.positionComponents.Components {
		self.ctx.Call("save")
		circle := (*self.circleComponents.Components[entity]).(*CircleRendererComponent)
		position := (*self.positionComponents.Components[entity]).(*game.PositionComponent)

		vector := self.lerpingComponents[entity]
		x := vector.X()
		y := vector.Y()

		X := math.Lerp(x, float64(position.Position.X()), 0.25)
		Y := math.Lerp(y, float64(position.Position.Y()), 0.25)
		self.lerpingComponents[entity] = math.NewVector(X, Y)

		color, e := circle.Color.Value()

		if e == nil {
			self.ctx.Set("fillStyle", color)
		}

		self.ctx.Call("translate", X, Y)
		self.ctx.Call("beginPath")
		self.ctx.Call("arc", 0, 0, circle.Radius, 0, 2*math2.Pi)
		self.ctx.Call("fill")
		self.ctx.Call("restore")
	}

}

type CircleRendererComponent struct {
	Size   math.Vector
	Color  MyHexColor
	Radius float32
}

func (self *CircleRendererComponent) Id() int {
	return int(CircleComponentType)
}

func (self *CircleRendererComponent) CreateComponent() {

}

func (self *CircleRendererComponent) DestroyComponent() {

}

func (self *CircleRendererComponent) Clone() Component {
	component := new(CircleRendererComponent)
	component.Size = self.Size
	component.Color = self.Color
	component.Radius = self.Radius
	return component
}

func (self *CircleRendererComponent) Reset(component Component) {

}

type MyHexColor Color

type errUnsupportedType struct {
	got  interface{}
	want reflect.Type
}

func (hc *MyHexColor) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errUnsupportedType{got: reflect.TypeOf(value), want: reflect.TypeOf("")}
	}
	c, err := Hex(s)
	if err != nil {
		return err
	}
	*hc = MyHexColor(c)
	return nil
}

func (hc *MyHexColor) Value() (driver.Value, error) {
	return Color(*hc).Hex(), nil
}

func (e errUnsupportedType) Error() string {
	return fmt.Sprintf("unsupported type: got %v, want a %s", e.got, e.want)
}

func (self *MyHexColor) UnmarshalJSON(bytes []byte) error {
	data := ""

	err := json.Unmarshal(bytes, &data)

	if err != nil {
		return err
	}

	color, err := Hex(data)

	if err != nil {
		return err
	}

	self.R = color.R
	self.G = color.G
	self.B = color.B

	return nil
}
