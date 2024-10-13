package frames

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/render"
)

// Game represents the Frames game state and logic
type Game struct {
	renderer *render.Renderer
	logger   *slog.Logger
	fps      float64
	lastTime time.Time
}

// NewGame creates a new instance of the Frames game
func NewGame(renderer *render.Renderer, logger *slog.Logger) *Game {
	return &Game{
		renderer: renderer,
		logger:   logger,
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
	g.renderer.DrawText(fpsText, 1, 1)
	g.renderer.Render()
}

// HandleInput processes user input
func (g *Game) HandleInput(input core.InputEvent) error {
	if  input.Key == core.KeyBackspace {
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
