package space_invaders

import "github.com/kuhree/gg/internal/engine/objects"

// Object represents a basic game entity
type Object struct {
	objects.GameObject

	Speed    objects.Vector2D
	Health    float64
	MaxHealth float64
}

func (o *Object) Size() float64 {
	return o.Height * o.Width
}

// Player represents the player's ship
type Player struct {
	Object

	Lives int
}

type AlienType int

// Alien represents an enemy alien
type Alien struct {
	Object

	AlienType     AlienType
	shootCooldown float64
	shootInterval float64
	shootChance   float64
}

// Projectile represents a projectile fired by the player or aliens
type Projectile struct {
	Object

	Source *Object
}

// Barrier represents a defensive structure
type Barrier struct {
	Object

	RegenerationRate float64
}

type CollectableType int

type Collectable struct {
	Object
	CollectableType CollectableType
	Duration        float64
}
