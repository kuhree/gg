package space_invaders

// Vector2D represents a 2D position or velocity
type Vector2D struct {
	X, Y float64
}

// GameObject represents a basic game entity
type GameObject struct {
	Position Vector2D
	Speed    Vector2D
	Height   float64
	Width    float64

	Health    float64
	MaxHealth float64
}

func (o *GameObject) Size() float64 {
	return o.Height * o.Width
}

// Player represents the player's ship
type Player struct {
	GameObject

	Lives int
}

type AlienType int

// Alien represents an enemy alien
type Alien struct {
	GameObject

	AlienType     AlienType
	shootCooldown float64
	shootInterval float64
	shootChance   float64
}

// Projectile represents a projectile fired by the player or aliens
type Projectile struct {
	GameObject

	Source *GameObject
}

// Barrier represents a defensive structure
type Barrier struct {
	GameObject

	RegenerationRate float64
}

type CollectableType int

type Collectable struct {
	GameObject
	CollectableType CollectableType
	Duration        float64
}
