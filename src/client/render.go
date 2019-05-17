package client

import (
	"fmt"
	"github.com/goburrow/dynamic"
	"github.com/lucasb-eyer/go-colorful"
	"io-engine-backend/src/game"
	"io-engine-backend/src/math"
	. "io-engine-backend/src/shared"
	math2 "math"
	"syscall/js"
)

type CanvasRenderSystem struct {
	circleComponents Storage
	positionComponents Storage
}

func (self *CanvasRenderSystem) Init() {
	dynamic.Register("CircleRendererComponent", func() interface{} {
		return &CircleRendererComponent{}
	})

	dynamic.Register("RenderGlobal", func() interface{} {
		return &RenderGlobal{}
	})

	self.circleComponents = NewStorage()
	self.positionComponents = NewStorage()
}

func (self *CanvasRenderSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{PositionComponentType, CircleComponentType}
}

func (self *CanvasRenderSystem) UpdateFrequency() int {
	return 1
}

func (self *CanvasRenderSystem) AddToStorage(entity Entity) {
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

	r := world.Globals[RenderGlobalType].(*RenderGlobal)

	r.ctx.Call("clearRect", 0, 0, r.Width, r.Height)

	for entity, _ := range self.positionComponents.Components {
		r.ctx.Call("save")
		circle := (*self.circleComponents.Components[entity]).(*CircleRendererComponent)
		position := (*self.positionComponents.Components[entity]).(*game.PositionComponent)

		color, e := circle.Color.Value()

		if e == nil {
			r.ctx.Set("fillStyle", color)
		}

		r.ctx.Call("translate", position.Position.X(), position.Position.Y())
		r.ctx.Call("beginPath")
		r.ctx.Call("arc", 0, 0, circle.Radius, 0, 2 * math2.Pi)
		r.ctx.Call("fill")
		r.ctx.Call("restore")
	}

}

type RenderGlobal struct {
	CanvasElementId string
	Width int
	Height int
	Scale math.Vector

	doc js.Value
	canvasElement js.Value
	ctx js.Value

	ShouldRender bool
}

func (self *RenderGlobal) Id() int {
	return int(RenderGlobalType)
}

func (self *RenderGlobal) CreateGlobal(world *World) {

	self.doc = js.Global().Get("document")

	if self.doc.Truthy() {
		fmt.Println("Setting Up Canvas Renderer on element: ", self.CanvasElementId)

		self.canvasElement = js.Global().Get("document").Call("getElementById", self.CanvasElementId)
		self.canvasElement.Set("width", self.Width)
		self.canvasElement.Set("height", self.Height)
		self.ctx = self.canvasElement.Call("getContext", "2d")
	}

}


type CircleRendererComponent struct {
	Size math.Vector
	Color colorful.HexColor
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
	return new(CircleRendererComponent)
}




