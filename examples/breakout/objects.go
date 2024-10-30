package breakout

import (
	"github.com/kuhree/gg/internal/engine/objects"
	"github.com/kuhree/gg/internal/engine/render"
)

// Paddle represents the player-controlled paddle
type Paddle struct {
	objects.GameObject
	Speed float64
}

// Ball represents the bouncing ball
type Ball struct {
	objects.GameObject
	Velocity objects.Vector2D
	Attached bool // When true, ball moves with paddle before launch
}

// Brick represents a destructible brick
type Brick struct {
	objects.GameObject
	Health int
	Points int
	Color  render.Color
}
