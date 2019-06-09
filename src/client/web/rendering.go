package web

import (
	"fmt"
	"github.com/goburrow/dynamic"
	"github.com/lucasb-eyer/go-colorful"
	. "io-engine-backend/src/ecs"
	"io-engine-backend/src/game"
	"io-engine-backend/src/math"
	math2 "math"
	"syscall/js"
)

type CanvasRenderSystem struct {
	circleComponents   Storage
	positionComponents Storage

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
		int(CircleComponentType): &self.circleComponents,
	}
	RemoveComponentsFromStorage(entity, storages)
}


func (self *CanvasRenderSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{PositionComponentType, CircleComponentType}
}

func (self *CanvasRenderSystem) UpdateFrequency() int {
	return 1
}

func (self *CanvasRenderSystem) AddToStorage(entity *Entity) {
	for k := range entity.Components {
		component := entity.Components[k].(Component)

		if component.Id() == int(PositionComponentType) {
			self.positionComponents.Components[entity.Id] = &component
		} else if component.Id() == int(CircleComponentType) {
			self.circleComponents.Components[entity.Id] = &component
		}
	}
}

func (self *CanvasRenderSystem) UpdateSystem(delta float64, world *World) {

	self.ctx.Call("clearRect", 0, 0, self.Width, self.Height)

	for entity, _ := range self.positionComponents.Components {
		self.ctx.Call("save")
		circle := (*self.circleComponents.Components[entity]).(*CircleRendererComponent)
		position := (*self.positionComponents.Components[entity]).(*game.PositionComponent)

		color, e := circle.Color.Value()

		if e == nil {
			self.ctx.Set("fillStyle", color)
		}

		self.ctx.Call("translate", position.Position.X(), position.Position.Y())
		self.ctx.Call("beginPath")
		self.ctx.Call("arc", 0, 0, circle.Radius, 0, 2*math2.Pi)
		self.ctx.Call("fill")
		self.ctx.Call("restore")
	}

}

type CircleRendererComponent struct {
	Size   math.Vector
	Color  colorful.HexColor
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
