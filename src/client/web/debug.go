package web

import (
	"encoding/json"
	"fmt"
	"github.com/Banyango/io-engine/src/ecs"
	"github.com/Banyango/io-engine/src/math"
	"reflect"
	"syscall/js"
)

type DebugSystem struct {
	CachedNumOfEntites int
	domParent          js.Value
	Entities           []*ecs.Entity
	delta              float64
}

func (self *DebugSystem) Init(w *ecs.World) {
	self.domParent = js.Global().Get("document").Call("getElementById", "debugSection")
}

func (self *DebugSystem) AddToStorage(entity *ecs.Entity) {

}

func (self *DebugSystem) RemoveFromStorage(entity *ecs.Entity) {
	fmt.Println("Removing ", entity.Id)
	info := map[string]interface{}{
		"type":"REMOVE_ALL_ENTITY",
	}
	marshal, _ := json.Marshal(info)
	self.Entities = nil

	js.Global().Get("window").Get("store").Call("dispatch", js.Global().Get("JSON").Call("parse", string(marshal)))
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
	if len(world.Entities) != len(self.Entities) {
		self.UpdateList(world)
		self.delta = 0
	}

	self.UpdateWorldInfo(world)

	if self.delta > 0.5 {
		self.UpdateComponents(world)
		self.delta = 0
	} else {
		self.delta += delta
	}
}

func (self *DebugSystem) createChildElement(element string, className string, innerHTML string) js.Value {
	value := js.Global().Get("document").Call("createElement", element)
	value.Set("className", className)
	value.Set("innerHTML", innerHTML)
	return value
}

func (self *DebugSystem) UpdateComponents(world *ecs.World) {
	for _, val := range world.Entities {
		for i := range val.Components {
			c := self.generateUpdateForComponent(val.Id, val.Components[i])
			marshal, _ := json.Marshal(dispatchComponent{
				Type:    "UPDATE_ENTITY",
				Id:      fmt.Sprint(val.Id),
				Payload: c,
			})

			js.Global().Get("window").Get("store").Call("dispatch", js.Global().Get("JSON").Call("parse", string(marshal)))
		}
	}
}

func (self *DebugSystem) generateUpdateForComponent(entityId int64, comp ecs.Component) components {

	propVal := propValue{}

	s := reflect.ValueOf(comp).Elem()
	typeOfT := s.Type()

	fields := []propValue{}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		if f.CanInterface() {
			if cast, ok := f.Interface().(math.Vector); ok {
				propVal.Name = typeOfT.Field(i).Name
				propVal.Value = vecField{
					T: "vector",
					X: fmt.Sprint(cast.X()),
					Y: fmt.Sprint(cast.Y()),
				}
				fields = append(fields, propVal)
			}
			if cast, ok := f.Interface().(math.VectorInt); ok {
				propVal.Name = typeOfT.Field(i).Name
				propVal.Value = vecField{
					T: "vector",
					X: fmt.Sprint(cast.X()),
					Y: fmt.Sprint(cast.Y()),
				}
				fields = append(fields, propVal)
			}
			if cast, ok := f.Interface().(float64); ok {
				propVal.Name = typeOfT.Field(i).Name
				propVal.Value = floatField{
					T:   "float",
					Val: fmt.Sprint(cast),
				}
				fields = append(fields, propVal)
			}
			if cast, ok := f.Interface().(float32); ok {
				propVal.Name = typeOfT.Field(i).Name
				propVal.Value = floatField{
					T:   "float",
					Val: fmt.Sprint(cast),
				}
				fields = append(fields, propVal)
			}
			if cast, ok := f.Interface().(int16); ok {
				propVal.Name = typeOfT.Field(i).Name
				propVal.Value = floatField{
					T:   "float",
					Val: fmt.Sprint(cast),
				}
				fields = append(fields, propVal)
			}
			if cast, ok := f.Interface().(int32); ok {
				propVal.Name = typeOfT.Field(i).Name
				propVal.Value = floatField{
					T:   "float",
					Val: fmt.Sprint(cast),
				}
				fields = append(fields, propVal)
			}
		}
	}

	return components{
		Id:         fmt.Sprint(entityId, comp.Id()),
		Name:       typeOfT.Name(),
		Properties: fields,
	}
}

func (self *DebugSystem) UpdateList(world *ecs.World) {

	for _, val := range world.Entities {
		if !self.EntitiesContains(val) {
			comps := make([]components, 0)

			for i := range val.Components {
				comps = append(comps, self.generateUpdateForComponent(val.Id, val.Components[i]))
			}

			marshal, _ := json.Marshal(dispatch{
				Type: "ADD_ENTITY",
				Payload: entityDispatch{
					Id:         fmt.Sprint(val.Id),
					Name:       val.Name,
					Components: comps,
				},
			})

			js.Global().Get("window").Get("store").Call("dispatch", js.Global().Get("JSON").Call("parse", string(marshal)))
		}
	}

	self.Entities = nil
	for _, val := range world.Entities {
		self.Entities = append(self.Entities, val)
	}
}

func (self *DebugSystem) EntitiesContains(entity *ecs.Entity) bool {
	for _, val := range self.Entities {
		if val.Id == entity.Id {
			return true
		}
	}
	return false
}

func (self *DebugSystem) UpdateWorldInfo(world *ecs.World) {
	info := map[string]interface{}{
		"type":"UPDATE_WORLD_INFO",
		"payload":map[string]interface{} {
			"currentTick":world.CurrentTick,
			"lastServerTick":world.LastServerTick,
			"ping":world.Ping,
			"bytesRec":world.BytesRec,
			"rttTicks":world.Ping/ecs.FIXED_DELTA,
		},
	}
	marshal, _ := json.Marshal(info)
	js.Global().Get("window").Get("store").Call("dispatch", js.Global().Get("JSON").Call("parse", string(marshal)))
}

type dispatchNoPayload struct {
	Type string `json:"type"`
}

type dispatchComponent struct {
	Type    string     `json:"type"`
	Id      string     `json:"id"`
	Payload components `json:"payload"`
}
type dispatch struct {
	Type    string         `json:"type"`
	Payload entityDispatch `json:"payload"`
}

type propValue struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type vecField struct {
	T string `json:"type"`
	X string
	Y string
}

type floatField struct {
	T   string `json:"type"`
	Val string
}

type components struct {
	Id         string      `json:"id"`
	Name       string      `json:"name"`
	Properties []propValue `json:"properties"`
}

type entityDispatch struct {
	Id         string       `json:"id"`
	Name       string       `json:"name"`
	Components []components `json:"components"`
}

type componentDispatch struct {
	Id         string     `json:"entityId"`
	Components components `json:"component"`
}
