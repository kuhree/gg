package sorts

import (
	"fmt"
	"math/rand"

	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/render"
)

type BaseScene struct {
	*Game
	sceneName     string
	blinkTimer    float64
	blinkInterval float64
	showOnBlink   bool
}

func (s *BaseScene) Enter() {
	s.Logger.Info("Entering scene", "scene", s.sceneName)
}

func (s *BaseScene) Exit() {
	s.Logger.Info("Exiting scene", "scene", s.sceneName)
}

func (s *BaseScene) Update(dt float64) {
	s.blinkTimer += dt
	if s.blinkTimer >= s.blinkInterval {
		s.blinkTimer = 0
		s.showOnBlink = !s.showOnBlink
	}
}

func (s *BaseScene) HandleInput(input core.InputEvent) error {
	return nil
}

type MainMenuScene struct {
	BaseScene
}

func NewMainMenuScene(game *Game) *MainMenuScene {
	return &MainMenuScene{
		BaseScene: BaseScene{
			Game:          game,
			sceneName:     "Main Menu",
			blinkInterval: 0.5,
			showOnBlink:   true,
		},
	}
}

func (s *MainMenuScene) Draw(renderer *render.Renderer) {
	width, height := s.Size()
	startX := width / 10

	_ = renderer.DrawText(fmt.Sprintf("%s - %s", s.Config.Title, s.sceneName), startX, int(float64(height)*s.Config.TitleOffset), render.ColorWhite)

	if s.showOnBlink {
		_ = renderer.DrawText("Press ENTER to start", startX, int(float64(height)*s.Config.ScoreOffset), render.ColorBrightMagenta)
	}

	controlsY := int(float64(height) * s.Config.ControlsOffset)
	_ = renderer.DrawText("Controls:", startX, controlsY, render.ColorBlue)
	_ = renderer.DrawText("1-3: Select sorting algorithm", startX, controlsY+s.Config.LineSpacing, render.ColorWhite)
	_ = renderer.DrawText("R: Reset array", startX, controlsY+2*s.Config.LineSpacing, render.ColorWhite)
	_ = renderer.DrawText("SPACE: Start/Pause sort", startX, controlsY+3*s.Config.LineSpacing, render.ColorWhite)
	_ = renderer.DrawText("Q: Quit", startX, controlsY+4*s.Config.LineSpacing, render.ColorWhite)
}

func (s *MainMenuScene) HandleInput(input core.InputEvent) error {
	switch input.Rune {
	case core.KeyEnter:
		s.Scenes.ChangeScene(VisualizerSceneID)
	case 'q', 'Q':
		return core.ErrQuitGame
	}
	return nil
}

type VisualizerScene struct {
	BaseScene
	updateTimer float64
	isPaused    bool
}

func NewVisualizerScene(game *Game) *VisualizerScene {
	scene := &VisualizerScene{
		BaseScene: BaseScene{
			Game:          game,
			sceneName:     "Visualizer",
			blinkInterval: 0.5,
			showOnBlink:   true,
		},
		isPaused: true,
	}
	scene.resetArray()
	scene.CurrentSorter = NewQuickSort()
	return scene
}

func (s *VisualizerScene) resetArray() {
	s.CurrentArray = make([]int, s.Config.ArraySize)
	for i := range s.CurrentArray {
		s.CurrentArray[i] = rand.Intn(s.Config.MaxValue) + 1
	}
	s.ComparisonCount = 0
	s.SwapCount = 0
	s.ElapsedTime = 0
	s.SortComplete = false
}

func (s *VisualizerScene) Update(dt float64) {
	s.BaseScene.Update(dt)
	
	if !s.isPaused && !s.SortComplete {
		s.updateTimer += dt
		s.ElapsedTime += dt
		
		if s.updateTimer >= s.Config.UpdateInterval {
			s.updateTimer = 0
			s.SortComplete = s.CurrentSorter.Step(s.CurrentArray)
		}
	}
}

func (s *VisualizerScene) Draw(renderer *render.Renderer) {
	width, height := s.Size()
	startX := width / 10

	// Draw title and status
	_ = renderer.DrawText(fmt.Sprintf("%s - %s", s.Config.Title, s.sceneName), startX, int(float64(height)*s.Config.TitleOffset), render.ColorWhite)
	
	// Draw array visualization
	maxHeight := height - 10
	for i, val := range s.CurrentArray {
		barHeight := int(float64(val) / float64(s.Config.MaxValue) * float64(maxHeight))
		x := startX + int(float64(i)*s.Config.BarWidth)
		
		for y := 0; y < barHeight; y++ {
			_ = renderer.DrawChar('█', x, height-5-y, render.ColorCyan)
		}
	}

	// Draw statistics
	statsY := height - 3
	_ = renderer.DrawText(fmt.Sprintf("Algorithm: %s", s.CurrentSorter.Name()), startX, statsY, render.ColorWhite)
	_ = renderer.DrawText(fmt.Sprintf("Comparisons: %d", s.ComparisonCount), startX+30, statsY, render.ColorWhite)
	_ = renderer.DrawText(fmt.Sprintf("Swaps: %d", s.SwapCount), startX+50, statsY, render.ColorWhite)
	_ = renderer.DrawText(fmt.Sprintf("Time: %.2fs", s.ElapsedTime), startX+70, statsY, render.ColorWhite)

	if s.SortComplete {
		_ = renderer.DrawText("Sort Complete!", startX, statsY-1, render.ColorGreen)
	}
}

func (s *VisualizerScene) HandleInput(input core.InputEvent) error {
	switch input.Rune {
	case '1':
		s.CurrentSorter = NewQuickSort()
		s.resetArray()
	case '2':
		s.CurrentSorter = NewBubbleSort()
		s.resetArray()
	case '3':
		s.CurrentSorter = NewMergeSort()
		s.resetArray()
	case 'r', 'R':
		s.resetArray()
	case ' ':
		s.isPaused = !s.isPaused
	case 'q', 'Q':
		s.Scenes.ChangeScene(MainMenuSceneID)
	}
	return nil
}
