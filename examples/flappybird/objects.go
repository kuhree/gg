package flappybird

import (
	"github.com/kuhree/gg/internal/engine/objects"
	"github.com/kuhree/gg/internal/engine/render"
)

// Bird represents the player-controlled bird
type Bird struct {
	objects.GameObject
	Velocity    float64
	Gravity     float64
	JumpForce   float64
	Character   rune
	Color       render.Color
	IsDead      bool
}

// Pipe represents an obstacle pipe
type Pipe struct {
	objects.GameObject
	Color       render.Color
	IsUpperPipe bool
}

// NewBird creates a new bird instance
func NewBird(x, y float64, config *Config) *Bird {
	return &Bird{
		GameObject: objects.GameObject{
			Position: objects.Vector2D{X: x, Y: y},
			Width:    1,
			Height:   1,
		},
		Velocity:    0,
		Gravity:     config.BirdGravity,
		JumpForce:   config.BirdJumpForce,
		Character:   '>', 
		Color:       render.ColorYellow,
		IsDead:      false,
	}
}

// NewPipe creates a new pipe instance
func NewPipe(x, y float64, height float64, isUpper bool) *Pipe {
	return &Pipe{
		GameObject: objects.GameObject{
			Position: objects.Vector2D{X: x, Y: y},
			Width:    2,
			Height:   height,
		},
		Color:       render.ColorGreen,
		IsUpperPipe: isUpper,
	}
}

