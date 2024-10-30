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
	
	// Center box calculations
	boxWidth := 50
	boxHeight := 12
	boxStartX := (width - boxWidth) / 2
	boxStartY := (height - boxHeight) / 2
	
	// Draw decorative box
	for y := 0; y < boxHeight; y++ {
		for x := 0; x < boxWidth; x++ {
			char := ' '
			color := render.ColorBlack
			
			// Draw borders
			if y == 0 || y == boxHeight-1 {
				if x == 0 || x == boxWidth-1 {
					char = '+'
				} else {
					char = '-'
				}
				color = render.ColorBlue
			} else if x == 0 || x == boxWidth-1 {
				char = '|'
				color = render.ColorBlue
			}
			
			_ = renderer.DrawChar(char, boxStartX+x, boxStartY+y, color)
		}
	}
	
	// Draw title centered in box
	title := fmt.Sprintf("%s - %s", s.Config.Title, s.sceneName)
	titleX := boxStartX + (boxWidth-len(title))/2
	_ = renderer.DrawText(title, titleX, boxStartY+2, render.ColorWhite)
	
	// Draw blinking start message
	if s.showOnBlink {
		startMsg := "Press ENTER to start"
		startX := boxStartX + (boxWidth-len(startMsg))/2
		_ = renderer.DrawText(startMsg, startX, boxStartY+4, render.ColorBrightMagenta)
	}
	
	// Draw controls section
	controlsY := boxStartY + 6
	controlsX := boxStartX + 3
	_ = renderer.DrawText("Controls:", controlsX, controlsY, render.ColorBlue)
	_ = renderer.DrawText("1-3: Select sorting algorithm", controlsX, controlsY+1, render.ColorWhite)
	_ = renderer.DrawText("R: Reset array", controlsX, controlsY+2, render.ColorWhite)
	_ = renderer.DrawText("SPACE: Start/Pause sort", controlsX, controlsY+3, render.ColorWhite)
	_ = renderer.DrawText("Q: Quit", controlsX, controlsY+4, render.ColorWhite)
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
			s.SortComplete = s.CurrentSorter.Step(s.CurrentArray, s.Game)
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
		// Draw completion stats box with padding and outline
		boxWidth := 44  // Increased width for padding
		boxHeight := 10  // Increased height for padding
		boxStartY := height/2 - boxHeight/2
		boxStartX := width/2 - boxWidth/2
		
		// Draw white outline
		for y := 0; y < boxHeight; y++ {
			for x := 0; x < boxWidth; x++ {
				if y == 0 || y == boxHeight-1 || x == 0 || x == boxWidth-1 {
					_ = renderer.DrawChar('█', boxStartX+x, boxStartY+y, render.ColorWhite)
				} else {
					_ = renderer.DrawChar(' ', boxStartX+x, boxStartY+y, render.ColorBlack)
				}
			}
		}
		
		// Draw completion stats with padding
		_ = renderer.DrawText("Sort Complete!", boxStartX+3, boxStartY+2, render.ColorWhite)
		_ = renderer.DrawText(fmt.Sprintf("Algorithm: %s", s.CurrentSorter.Name()), boxStartX+3, boxStartY+3, render.ColorWhite)
		_ = renderer.DrawText(fmt.Sprintf("Comparisons: %d", s.ComparisonCount), boxStartX+3, boxStartY+4, render.ColorWhite)
		_ = renderer.DrawText(fmt.Sprintf("Swaps: %d", s.SwapCount), boxStartX+3, boxStartY+5, render.ColorWhite)
		_ = renderer.DrawText(fmt.Sprintf("Time: %.2fs", s.ElapsedTime), boxStartX+3, boxStartY+6, render.ColorWhite)
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
