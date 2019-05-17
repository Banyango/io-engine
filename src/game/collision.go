package game

import (
	"github.com/SolarLune/resolv/resolv"
	"github.com/goburrow/dynamic"
	"io-engine-backend/src/math"
	. "io-engine-backend/src/shared"
)


/*
----------------------------------------------------------------------------------------------------------------
Collision System
----------------------------------------------------------------------------------------------------------------
*/


type CollisionSystem struct {
	positionComponents  Storage
	collisionComponents Storage
	space *resolv.Space
}

func (self *CollisionSystem) Init() {
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

func (self *CollisionSystem) AddToStorage(entity Entity) {
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

		collider.Reset()
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

func (self *PositionComponent) Id() int {
	return int(PositionComponentType)
}

func (self *PositionComponent) CreateComponent() {

}

func (self *PositionComponent) DestroyComponent() {

}

func (self *PositionComponent) Clone() Component {
	return new(PositionComponent)
}

type CollisionComponent struct {
	Size     math.VectorInt `json:"size"`

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
	return new(CollisionComponent)
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
	c.shape = resolv.NewRectangle(int32(0),int32(0), int32(c.Size.X()), int32(c.Size.Y()))
}

func (c *CollisionComponent) DestroyComponent() {

}


func (c *CollisionComponent) Extents(position math.VectorInt) (math.VectorInt, math.VectorInt) {
	return c.Min(position), c.Max(position)
}

func (c *CollisionComponent) Reset() {
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
