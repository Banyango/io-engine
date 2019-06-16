package game

import (
	"github.com/SolarLune/resolv/resolv"
	"github.com/goburrow/dynamic"
	. "io-engine-backend/src/ecs"
	"io-engine-backend/src/math"
	"io-engine-backend/src/server"
	math2 "math"
)

/*
----------------------------------------------------------------------------------------------------------------
Collision System
----------------------------------------------------------------------------------------------------------------
*/

type CollisionSystem struct {
	positionComponents  Storage
	collisionComponents Storage
	space               *resolv.Space
}

func (self *CollisionSystem) Init(w *World) {
	dynamic.Register("PositionComponent", func() interface{} {
		return &PositionComponent{}
	})

	dynamic.Register("CollisionComponent", func() interface{} {
		return &CollisionComponent{}
	})

	self.positionComponents = NewStorage()
	self.collisionComponents = NewStorage()

	self.space = resolv.NewSpace()
}

func (self *CollisionSystem) RequiredComponentTypes() []ComponentType {
	return []ComponentType{PositionComponentType, CollisionComponentType}
}

func (self *CollisionSystem) UpdateFrequency() int {
	return 60
}

func (self *CollisionSystem) RemoveFromStorage(entity *Entity) {
	storages := map[int]*Storage{
		int(PositionComponentType):  &self.positionComponents,
		int(CollisionComponentType): &self.collisionComponents,
	}
	RemoveComponentsFromStorage(entity, storages)
}

func (self *CollisionSystem) AddToStorage(entity *Entity) {
	for k := range entity.Components {
		component := entity.Components[k].(Component)

		if component.Id() == int(PositionComponentType) {
			self.positionComponents.Components[entity.Id] = &component
		} else if component.Id() == int(CollisionComponentType) {
			self.collisionComponents.Components[entity.Id] = &component

			collider := (*self.collisionComponents.Components[entity.Id]).(*CollisionComponent)

			self.space.Add(collider.shape)
		}
	}
}

func (self *CollisionSystem) UpdateSystem(delta float64, world *World) {

	for entity, _ := range self.collisionComponents.Components {
		position := (*self.positionComponents.Components[entity]).(*PositionComponent)
		collider := (*self.collisionComponents.Components[entity]).(*CollisionComponent)

		// make sure the position is equal to the collider position
		collider.shape.X = int32(position.Position.X())
		collider.shape.Y = int32(position.Position.Y())

		collider.ResetBooleans()
	}

	for entity, _ := range self.collisionComponents.Components {
		position := (*self.positionComponents.Components[entity]).(*PositionComponent)
		collider := (*self.collisionComponents.Components[entity]).(*CollisionComponent)

		totalVelocity := collider.Velocity.Scale(delta).Add(collider.Remaining)

		velocityRoundedToPixel := totalVelocity.Truncate().ToInt()
		collider.Remaining = totalVelocity.Remaining()

		if velocityRoundedToPixel.X() > 0 && velocityRoundedToPixel.Y() > 0 {
			for otherEntityId, _ := range self.collisionComponents.Components {
				if entity != otherEntityId {
					otherCollider := (*self.collisionComponents.Components[otherEntityId]).(*CollisionComponent)
					res := resolv.Resolve(collider.shape, otherCollider.shape, int32(velocityRoundedToPixel.X()), int32(velocityRoundedToPixel.Y()))
					if (res.Colliding()) {
						collider.AddEntityToCollisionList(otherEntityId)
						otherCollider.AddEntityToCollisionList(entity)
						//collision = true
					}
				}
			}
		}

		//if !collision {
		position.Position = position.Position.Add(velocityRoundedToPixel)
		//collider.shape.X = int32(position.Position.X())
		//collider.shape.Y = int32(position.Position.Y())
		//}

	}
}

func (self *CollisionSystem) Collides(
	min math.VectorInt,
	max math.VectorInt,
	otherMin math.VectorInt,
	otherMax math.VectorInt) bool {
	if min.X() <= otherMin.X() ||
		min.X() >= otherMax.X() ||
		max.Y() <= otherMin.Y() ||
		min.Y() >= otherMax.Y() {
		return false
	}
	return true
}

/*
----------------------------------------------------------------------------------------------------------------
Position Component
----------------------------------------------------------------------------------------------------------------
*/

type PositionComponent struct {
	Position math.VectorInt
}

func (self *PositionComponent) Reset(component Component) {
	if val, ok := component.(*PositionComponent); ok {
		self.Position.Set(val.Position.X(), val.Position.Y())
	}
}

func (self *PositionComponent) ReadUDP(networkPacket *server.NetworkData) {
	var data struct {
		X int
		Y int
	}

	server.DecodeNetworkDataBytes(networkPacket, self.Id(), &data)

	self.Position.Set(data.X, data.Y)
}

func (self *PositionComponent) WriteUDP(networkPacket *server.NetworkData) {
	var data struct {
		X int
		Y int
	}

	data.X = self.Position.X()
	data.Y = self.Position.Y()

	server.EncodeNetworkDataBytes(networkPacket, self.Id(), data)
}

func (self *PositionComponent) Id() int {
	return int(PositionComponentType)
}

func (self *PositionComponent) CreateComponent() {

}

func (self *PositionComponent) DestroyComponent() {

}

func (self *PositionComponent) AreEquals(component Component) bool {
	if val, ok := component.(*PositionComponent); ok {
		return val.Position.X() == self.Position.X() && val.Position.Y() == self.Position.Y()
	} else {
		return false
	}
}

func (self *PositionComponent) Clone() Component {
	component := new(PositionComponent)
	component.Position = self.Position
	return component
}

type CollisionComponent struct {
	Size math.VectorInt `json:"size"`

	Velocity math.Vector

	Remaining math.Vector

	entitiesCollidingWith []int64

	shape *resolv.Rectangle

	Bottom bool
	Top    bool
	Left   bool
	Right  bool

	WasBottom bool
	WasTop    bool
	WasLeft   bool
	WasRight  bool
}

func (c *CollisionComponent) Clone() Component {
	component := new(CollisionComponent)
	component.Size = c.Size
	component.Velocity = c.Velocity
	return component
}

func (c *CollisionComponent) AddEntityToCollisionList(entityId int64) {

	contains := false
	for _, i := range c.entitiesCollidingWith {
		if i == entityId {
			contains = true
		}
	}

	if !contains {
		c.entitiesCollidingWith = append(c.entitiesCollidingWith, entityId)
	}

}

func (c *CollisionComponent) Max(position math.VectorInt) math.VectorInt {
	return math.NewVectorInt(position.X()+c.Size.X()/2.0, position.Y()+c.Size.Y()/2)
}

func (c *CollisionComponent) Min(position math.VectorInt) math.VectorInt {
	return math.NewVectorInt(position.X()-c.Size.X()/2, position.Y()-c.Size.Y()/2)
}

func (c *CollisionComponent) Id() int {
	return int(CollisionComponentType);
}

func (c *CollisionComponent) CreateComponent() {
	c.shape = resolv.NewRectangle(int32(0), int32(0), int32(c.Size.X()), int32(c.Size.Y()))
}

func (c *CollisionComponent) DestroyComponent() {

}

func (self *CollisionComponent) Reset(component Component) {
	if val, ok := component.(*CollisionComponent); ok {
		self.Size.Set(val.Size.X(), val.Size.Y())
		self.Velocity.Set(val.Velocity.X(), val.Velocity.Y())
	}
}

func (c *CollisionComponent) Extents(position math.VectorInt) (math.VectorInt, math.VectorInt) {
	return c.Min(position), c.Max(position)
}

func (self *CollisionComponent) ReadUDP(networkPacket *server.NetworkData) {
	var data struct {
		VelX       float32
		VelY       float32
		RemainingX float32
		RemainingY float32
		//Collision byte
	}

	server.DecodeNetworkDataBytes(networkPacket, self.Id(), &data)

	self.Velocity.Set(float64(data.VelX), float64(data.VelY))
	self.Remaining.Set(float64(data.RemainingX), float64(data.RemainingY))
}

func (self *CollisionComponent) WriteUDP(networkPacket *server.NetworkData) {
	var data struct {
		VelX       float32
		VelY       float32
		RemainingX float32
		RemainingY float32
		//Collision byte
	}

	data.VelX = float32(self.Velocity.X())
	data.VelY = float32(self.Velocity.Y())
	data.RemainingX = float32(self.Remaining.X())
	data.RemainingY = float32(self.Remaining.Y())

	server.EncodeNetworkDataBytes(networkPacket, self.Id(), data)
}

func (self *CollisionComponent) AreEquals(component Component) bool {
	if val, ok := component.(*CollisionComponent); ok {
		return math2.Abs(val.Velocity.X()-self.Velocity.X()) < 0.001 &&
			math2.Abs(val.Velocity.Y()-self.Velocity.Y()) < 0.001 &&
			math2.Abs(val.Remaining.X()-self.Remaining.X()) < 0.001 &&
			math2.Abs(val.Remaining.Y()-self.Remaining.Y()) < 0.001
	} else {
		return false
	}
}

func (c *CollisionComponent) ResetBooleans() {
	c.WasBottom = c.Bottom
	c.WasLeft = c.Left
	c.WasRight = c.Right
	c.WasTop = c.Top

	c.Bottom = false
	c.Top = false
	c.Left = false
	c.Right = false

	c.entitiesCollidingWith = nil

}
