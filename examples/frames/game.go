package frames

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/render"
	"github.com/kuhree/gg/internal/utils"
)

// Game represents the Frames game state and logic
type Game struct {
	renderer *render.Renderer
	logger   *slog.Logger
	fps      float64
	lastTime time.Time
}

// NewGame creates a new instance of the Frames game
func NewGame(width, height int) *Game {
	renderer := render.NewRenderer(width, height) // Create a 80x24 ASCII renderer

	return &Game{
		renderer: renderer,
		logger:   utils.Logger,
		lastTime: time.Now(),
	}
}

// Update updates the game state
func (g *Game) Update(dt float64) error {
	now := time.Now()
	elapsed := now.Sub(g.lastTime).Seconds()
	g.fps = 1 / elapsed
	g.lastTime = now
	return nil
}

// Draw renders the game state
func (g *Game) Draw() {
	g.renderer.Clear()
	fpsText := fmt.Sprintf("FPS: %.2f", g.fps)
	_ = g.renderer.DrawText(fpsText, 1, 1, render.ColorBlue)
	g.renderer.Render()
}

// HandleInput processes user input
func (g *Game) HandleInput(input core.InputEvent) error {
	if input.Key == core.KeyBackspace || input.Rune == core.KeyQ {
		return core.ErrQuitGame
	}
	return nil
}

// Init initializes the game
func (g *Game) Init() error {
	g.logger.Info("Frames game initialized")
	return nil
}

// Cleanup performs any necessary cleanup
func (g *Game) Cleanup() {
	g.logger.Info("Frames game cleaned up")
}
