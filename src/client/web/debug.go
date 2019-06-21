package web

import (
	"github.com/Banyango/io-engine/src/ecs"
	"github.com/Banyango/io-engine/src/game"
	"strings"
	"syscall/js"
)

type DebugSystem struct {
	CachedNumOfEntites int
	domParent          js.Value
}

func (self *DebugSystem) Init(w *ecs.World) {
	self.domParent = js.Global().Get("document").Call("getElementById", "debugSection")
}

func (self *DebugSystem) AddToStorage(entity *ecs.Entity) {

}

func (self *DebugSystem) RemoveFromStorage(entity *ecs.Entity) {

}

func (self *DebugSystem) RequiredComponentTypes() []ecs.ComponentType {
	return []ecs.ComponentType{}
}

func (self *DebugSystem) createElement(element string, className string) js.Value {
	value := js.Global().Get("document").Call("createElement", element)
	value.Set("className", className)
	return value
}

func (self *DebugSystem) createCustomElement(element string, attributes map[string]string) js.Value {
	value := js.Global().Get("document").Call("createElement", element)

	for key, val := range attributes {
		value.Call("setAttribute", key, val)
	}

	return value
}

func (self *DebugSystem) UpdateSystem(delta float64, world *ecs.World) {
	if len(world.Entities) != self.CachedNumOfEntites {
		log("Recreating Entity List...")

		self.CachedNumOfEntites = len(world.Entities)

		ul := NewHTMLComponent("ul", "list-group","")

		log("Adding Entities...")

		for _, val := range world.Entities {

			div := NewHTMLComponent("div","list-group-item","")
			h4 := NewHTMLComponent("h4","mb-1", strings.Join([]string{"Entity:", string(val.Id)}, " "))

			h4.attachTo(div)

			ulComp := NewHTMLComponent("ul", "list-group", "")

			for _, val2 := range val.Components {
				if posComp, ok := val2.(*game.PositionComponent); ok {
					CreatePositionHTMLComponent(posComp.Position.X(), posComp.Position.Y()).attachTo(ulComp)
				}
			}

			ulComp.attachTo(div)
			div.attachTo(ul)

		}

		log("Appending HTML...")
		self.domParent.Set("innerHTML", "")
		self.domParent.Call("appendChild", ul.DomValue)
	}

}

// todo debug system for network state etc. import bootstrap.css and bind the ECS hierarchy to the dom on each update.
// todo only change the dom if there are entity changes.

type HTMLComponent struct {
	Tag string
	Class string
	InnerText string
	DomValue js.Value
}

func NewHTMLComponent(tag string, class string, innerText string) *HTMLComponent {
	component := HTMLComponent{Tag: tag, Class: class, InnerText:innerText}
	component.createHTML()
	return &component
}


func (self *HTMLComponent) createHTML() *HTMLComponent {
	self.DomValue = js.Global().Get("document").Call("createElement", self.Tag)
	self.DomValue.Set("className", self.Class)
	self.DomValue.Set("innerText", self.InnerText)
	return self
}

func (self *HTMLComponent) attachTo(parent *HTMLComponent) {
	parent.DomValue.Call("appendChild", self.DomValue)
}

func (self *DebugSystem) createChildElement(element string, className string, innerHTML string) js.Value {
	value := js.Global().Get("document").Call("createElement", element)
	value.Set("className", className)
	value.Set("innerHTML", innerHTML)
	return value
}

func CreatePositionHTMLComponent(x int, y int) *HTMLComponent {
	div := NewHTMLComponent("div","list-group-item","")

	h3 := NewHTMLComponent("h3", "","PositionComponent")
	//log("x=",x)
	p := NewHTMLComponent("p","","x="+string(x))
	p2 := NewHTMLComponent("p","","y="+string(y))

	h3.attachTo(div)
	p.attachTo(div)
	p2.attachTo(div)

	return div
}

