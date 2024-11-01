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

	lastTime   time.Time
	totalTime  float64
	frameCount int

	targetFps   float64
	currentFps  float64
	targetDelta float64

	// FPS stats for different intervals
	fpsStats5s  *fpsStats
	fpsStats10s *fpsStats
	fpsStats30s *fpsStats
}

// NewGame creates a new instance of the Frames game
func NewGame(width, height int, targetFps float64) *Game {
	renderer := render.NewRenderer(width, height, render.DefaultPalette)

	now := time.Now()
	return &Game{
		Width:       width,
		Height:      height,
		renderer:    renderer,
		logger:      utils.Logger,
		lastTime:    now,
		targetFps:   targetFps,
		targetDelta: 1.0 / targetFps,
		fpsStats5s:  newFpsStats(5 * time.Second),
		fpsStats10s: newFpsStats(10 * time.Second),
		fpsStats30s: newFpsStats(30 * time.Second),
	}
}

// Update updates the game state
func (g *Game) Update(dt float64) error {
	now := time.Now()
	elapsed := now.Sub(g.lastTime).Seconds()
	currentFps := 1 / elapsed
	g.currentFps = currentFps
	g.lastTime = now

	g.frameCount++
	g.totalTime += elapsed

	// Update interval stats
	g.fpsStats5s.update(currentFps, now)
	g.fpsStats10s.update(currentFps, now)
	g.fpsStats30s.update(currentFps, now)
	return nil
}

func (g *Game) Size() (int, int) {
	return g.Width, g.Height
}

// Draw renders the game state
func (g *Game) Draw() {
	g.renderer.Clear()

	// Display FPS info
	_ = g.renderer.DrawText(fmt.Sprintf("Target FPS: %.2f", g.targetFps), 2, 2, render.ColorWhite)
	_ = g.renderer.DrawText(fmt.Sprintf("Current FPS: %.2f", g.currentFps), 2, 3, render.ColorBlue)

	fpsDiff := g.currentFps - g.targetFps
	diffColor := render.ColorYellow
	if fpsDiff > 5 {
		diffColor = render.ColorGreen
	} else if fpsDiff < -5 {
		diffColor = render.ColorRed
	}
	_ = g.renderer.DrawText(fmt.Sprintf("FPS Diff: %+.2f", fpsDiff), 2, 4, diffColor)

	// Display FPS statistics for different intervals
	_ = g.renderer.DrawText("5 Second Stats:", 2, 6, render.ColorWhite)
	_ = g.renderer.DrawText(fmt.Sprintf("  Min: %.2f  Max: %.2f  Avg: %.2f", 
		g.fpsStats5s.min, g.fpsStats5s.max, g.fpsStats5s.avg()), 2, 7, render.ColorGreen)

	_ = g.renderer.DrawText("10 Second Stats:", 2, 9, render.ColorWhite)
	_ = g.renderer.DrawText(fmt.Sprintf("  Min: %.2f  Max: %.2f  Avg: %.2f", 
		g.fpsStats10s.min, g.fpsStats10s.max, g.fpsStats10s.avg()), 2, 10, render.ColorYellow)

	_ = g.renderer.DrawText("30 Second Stats:", 2, 12, render.ColorWhite)
	_ = g.renderer.DrawText(fmt.Sprintf("  Min: %.2f  Max: %.2f  Avg: %.2f", 
		g.fpsStats30s.min, g.fpsStats30s.max, g.fpsStats30s.avg()), 2, 13, render.ColorMagenta)

	// Display frame count and total time
	_ = g.renderer.DrawText(fmt.Sprintf("Frames: %d", g.frameCount), 2, 15, render.ColorCyan)
	_ = g.renderer.DrawText(fmt.Sprintf("Total Time: %.2fs", g.totalTime), 2, 16, render.ColorCyan)

	// Display window dimensions
	_ = g.renderer.DrawText(fmt.Sprintf("Window: %dx%d", g.Width, g.Height), 2, 18, render.ColorWhite)

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
// fpsStats tracks FPS statistics over a time interval
type fpsStats struct {
	interval time.Duration
	samples  []float64
	times    []time.Time
	min      float64
	max      float64
}

func newFpsStats(interval time.Duration) *fpsStats {
	return &fpsStats{
		interval: interval,
		min:      -1, // sentinel value
	}
}

func (s *fpsStats) update(fps float64, now time.Time) {
	// Remove old samples
	cutoff := now.Add(-s.interval)
	i := 0
	for i < len(s.times) && s.times[i].Before(cutoff) {
		i++
	}
	if i > 0 {
		s.samples = s.samples[i:]
		s.times = s.times[i:]
	}

	// Add new sample
	s.samples = append(s.samples, fps)
	s.times = append(s.times, now)

	// Update min/max
	s.min = fps
	s.max = fps
	for _, sample := range s.samples {
		if sample < s.min || s.min < 0 {
			s.min = sample
		}
		if sample > s.max {
			s.max = sample
		}
	}
}

func (s *fpsStats) avg() float64 {
	if len(s.samples) == 0 {
		return 0
	}
	sum := 0.0
	for _, fps := range s.samples {
		sum += fps
	}
	return sum / float64(len(s.samples))
}
