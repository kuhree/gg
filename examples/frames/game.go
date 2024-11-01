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
	Width  int
	Height int

	renderer *render.Renderer
	logger   *slog.Logger

	lastTime  time.Time
	totalTime float64

	frameCount int
	curremtFps float64
	minFps     float64
	maxFps     float64
	avgFps     float64
}

// NewGame creates a new instance of the Frames game
func NewGame(width, height int) *Game {
	renderer := render.NewRenderer(width, height, render.DefaultPalette) // Create a 80x24 ASCII renderer

	return &Game{
		Width:    width,
		Height:   height,
		renderer: renderer,
		logger:   utils.Logger,
		lastTime: time.Now(),
	}
}

// Update updates the game state
func (g *Game) Update(dt float64) error {
	now := time.Now()
	elapsed := now.Sub(g.lastTime).Seconds()
	currentFps := 1 / elapsed
	g.curremtFps = currentFps
	g.lastTime = now

	// Update stats
	g.frameCount++
	g.totalTime += elapsed

	if g.frameCount == 1 {
		g.minFps = currentFps
		g.maxFps = currentFps
	} else {
		if currentFps < g.minFps {
			g.minFps = currentFps
		}
		if currentFps > g.maxFps {
			g.maxFps = currentFps
		}
	}

	g.avgFps = float64(g.frameCount) / g.totalTime
	return nil
}

func (g *Game) Size() (int, int) {
	return g.Width, g.Height
}

// Draw renders the game state
func (g *Game) Draw() {
	g.renderer.Clear()

	// Display current FPS
	fpsText := fmt.Sprintf("Current FPS: %.2f", g.curremtFps)
	_ = g.renderer.DrawText(fpsText, 2, 2, render.ColorBlue)

	// Display FPS statistics
	_ = g.renderer.DrawText(fmt.Sprintf("Min FPS: %.2f", g.minFps), 2, 3, render.ColorGreen)
	_ = g.renderer.DrawText(fmt.Sprintf("Max FPS: %.2f", g.maxFps), 2, 4, render.ColorRed)
	_ = g.renderer.DrawText(fmt.Sprintf("Avg FPS: %.2f", g.avgFps), 2, 5, render.ColorYellow)

	// Display frame count and total time
	_ = g.renderer.DrawText(fmt.Sprintf("Frames: %d", g.frameCount), 2, 7, render.ColorCyan)
	_ = g.renderer.DrawText(fmt.Sprintf("Total Time: %.2fs", g.totalTime), 2, 8, render.ColorMagenta)

	// Display window dimensions
	_ = g.renderer.DrawText(fmt.Sprintf("Window: %dx%d", g.Width, g.Height), 2, 10, render.ColorWhite)

	g.renderer.Render()
}

// HandleInput processes user input
func (g *Game) HandleInput(input core.InputEvent) error {
	if input.Rune == core.KeyQ {
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
