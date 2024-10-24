package breakout

import "github.com/kuhree/gg/internal/engine/render"

// Vector2D represents a 2D position or velocity
type Vector2D struct {
	X, Y float64
}

// GameObject represents a basic game entity
type GameObject struct {
	Position Vector2D
	Height   float64
	Width    float64
}

// Paddle represents the player-controlled paddle
type Paddle struct {
	GameObject
	Speed float64
}

// Ball represents the bouncing ball
type Ball struct {
	GameObject
	Velocity Vector2D
	Attached bool // When true, ball moves with paddle before launch
}

// Brick represents a destructible brick
type Brick struct {
	GameObject
	Health int
	Points int
	Color  render.Color
}
