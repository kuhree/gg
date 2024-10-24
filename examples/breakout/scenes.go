package breakout

import (
	"fmt"

	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/leaderboard"
	"github.com/kuhree/gg/internal/engine/render"
)

const (
	titleOffset = 1.0 / 10
	lineSpacing = 2
)

// BaseScene provides common functionality for all scenes
type BaseScene struct {
	*Game
	sceneName     string
	blinkTimer    float64
	blinkInterval float64
	showOnBlink   bool
}

// Enter logs when a scene is entered
func (s *BaseScene) Enter() {
	s.Logger.Info("Entering scene", "scene", s.sceneName)
}

// Exit logs when a scene is exited
func (s *BaseScene) Exit() {
	s.Logger.Info("Exiting scene", "scene", s.sceneName)
}

// Update is a no-op for scenes that don't need updates
func (s *BaseScene) Update(dt float64) {
	s.blinkTimer += dt
	if s.blinkTimer >= s.blinkInterval {
		s.blinkTimer = 0
		s.showOnBlink = !s.showOnBlink
	}
}

// HandleInput is a no-op for scenes that don't handle input
func (s *BaseScene) HandleInput(input core.InputEvent) error {
	return nil
}

// MainMenuScene represents the main menu
type MainMenuScene struct {
	BaseScene
}

// PlayingScene represents the main gameplay
type PlayingScene struct {
	BaseScene
}

// PauseMenuScene represents the pause menu
type PauseMenuScene struct {
	BaseScene
}

// GameOverScene represents the game over screen
type GameOverScene struct {
	BaseScene
	name        string
	nameEntered bool
}

// NewMainMenuScene creates a new main menu scene
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

// NewPlayingScene creates a new playing scene
func NewPlayingScene(game *Game) *PlayingScene {
	scene := &PlayingScene{
		BaseScene: BaseScene{
			Game:          game,
			sceneName:     "Playing",
			blinkInterval: 0.5,
			showOnBlink:   true,
		},
	}

	return scene
}

// NewPauseMenuScene creates a new pause menu scene
func NewPauseMenuScene(game *Game) *PauseMenuScene {
	return &PauseMenuScene{
		BaseScene: BaseScene{
			Game:          game,
			sceneName:     "Pause Menu",
			blinkInterval: 0.5,
			showOnBlink:   true,
		},
	}
}

// NewGameOverScene creates a new game over scene
func NewGameOverScene(game *Game) *GameOverScene {
	return &GameOverScene{
		BaseScene: BaseScene{
			Game:          game,
			sceneName:     "Game Over",
			blinkInterval: 0.5,
			showOnBlink:   true,
		},
		name:        "",
		nameEntered: false,
	}
}

// MainMenuScene methods

func (s *MainMenuScene) Draw(renderer *render.Renderer) {
	width, height := s.Size()
	startX := width / 10

	const (
		startOffset    = 1.0 / 6
		controlsOffset = 2.0 / 8
	)

	_ = renderer.DrawText(fmt.Sprintf("%s - %s", s.Config.Title, s.sceneName), startX, int(float64(height)*titleOffset), render.ColorWhite)

	if s.showOnBlink {
		_ = renderer.DrawText("Press ENTER to start", startX, int(float64(height)*startOffset), render.ColorBrightMagenta)
	}

	controlsY := int(float64(height) * controlsOffset)
	_ = renderer.DrawText("Controls:", startX, controlsY, render.ColorBlue)
	_ = renderer.DrawText("ESC to pause", startX, controlsY+3*lineSpacing, render.ColorWhite)
	_ = renderer.DrawText("Q to pause/quit", startX, controlsY+4*lineSpacing, render.ColorWhite)
}

func (s *MainMenuScene) HandleInput(input core.InputEvent) error {
	switch input.Rune {
	case core.KeyEnter:
		s.Logger.Info("Starting new game")
		s.Scenes.ChangeScene(PlayingSceneID)
		return nil
	case 'q', 'Q':
		s.Scenes.ChangeScene(GameOverSceneID)
		return core.ErrQuitGame
	}

	return nil
}

// PlayingScene methods
func (s *PlayingScene) Update(dt float64) {
	s.BaseScene.Update(dt)

	s.updateCollisions(dt)
	ended, reason := s.checkGameState(dt)
	if ended {
		s.endGame(reason)
	}

	s.CurrentLevel++
}

func (s *PlayingScene) Draw(renderer *render.Renderer) {
	width, height := s.Size()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			s.drawObjOverlay(x, y, render.ColorWhite)
		}
	}

	// Draw score, level, lives...
	_ = renderer.DrawText(fmt.Sprintf("Score: %d", s.Score), 1, 1, render.ColorWhite)
	_ = renderer.DrawText(fmt.Sprintf("Level: %d", s.CurrentLevel), 1, 2, render.ColorWhite)
}

func (s *PlayingScene) HandleInput(input core.InputEvent) error {

	switch input.Rune {
	case '1', core.KeyF1:
		s.Debug = !s.Debug
	case '2', core.KeyF2:
		s.Overlay = !s.Overlay
	case 'p', 'P':
		s.Scenes.ChangeScene(PauseMenuSceneID)
	case 'q', 'Q':
		s.Scenes.ChangeScene(GameOverSceneID)
	}

	return nil
}

// PlayingScene helpers

// updateCollisions detects and handles collisions between game objects
func (s *PlayingScene) updateCollisions(_ float64) {
}

// checkGameState determines if the game should end
func (s *PlayingScene) checkGameState(_ float64)  (bool, string) {
	return false, ""
}

func (s *PlayingScene) endGame(reason string) {
	s.Logger.Info("Game over", "reason", reason, "score", s.Score, "level", s.CurrentLevel+1)
	s.Scenes.ChangeScene(GameOverSceneID)
}

func (s *PlayingScene) drawObjOverlay(x, y int, color render.Color) {
	if !s.Overlay && !s.Debug {
		return
	}

	if s.Overlay {
		char := '0'
		_ = s.Renderer.DrawChar(char, x, y, color)
	}

	if s.Debug {
		debugInfo := []string{}
		for i, info := range debugInfo {
			_ = s.Renderer.DrawText(info, x, y+i, color)
		}
	}
}

// PauseMenuScene methods

func (s *PauseMenuScene) Draw(renderer *render.Renderer) {
	const (
		scoreOffset    = 1.0 / 6
		controlsOffset = 1.0 / 4
	)

	width, height := s.Size()
	startX := width / 10

	// Draw title
	_ = renderer.DrawText(fmt.Sprintf("%s - %s", s.Config.Title, s.sceneName), startX, int(float64(height)*titleOffset), render.ColorWhite)

	if s.showOnBlink {
		_ = renderer.DrawText(
			fmt.Sprintf("Score: %d | Level: %d", s.Score, s.CurrentLevel),
			startX,
			int(float64(height)*scoreOffset),
			render.ColorMagenta,
		)
	}

	// Draw controls
	controlsY := int(float64(height) * controlsOffset)
	_ = renderer.DrawText("Controls:", startX, controlsY, render.ColorBlue)
	_ = renderer.DrawText("Press ESC to resume", startX, controlsY+lineSpacing, render.ColorWhite)
	_ = renderer.DrawText("Press Q to quit", startX, controlsY+2*lineSpacing, render.ColorWhite)
}

func (s *PauseMenuScene) HandleInput(input core.InputEvent) error {
	switch input.Rune {
	case core.KeyEscape:
		s.Scenes.ChangeScene(PlayingSceneID)
	case 'q', 'Q':
		s.Scenes.ChangeScene(GameOverSceneID)
		return core.ErrQuitGame
	}

	return nil
}

// GameOverScene methods

func (s *GameOverScene) Enter() {
	s.BaseScene.Enter()
	err := s.Leaderboard.Load(s.Config.BoardFile)
	if err != nil {
		s.Logger.Warn("Failed to load existing leaderboard. Creating a new one...", "path", s.Config.BoardFile, "err", err)
		s.Leaderboard.Records = make([]leaderboard.Record, 0)
	}
}

func (s *GameOverScene) GetDetails() string {
	width, height := s.Size()
	return fmt.Sprintf(
		"%dW*%dH|L%d",
		width, height,
		s.CurrentLevel,
	)
}

func (s *GameOverScene) Draw(renderer *render.Renderer) {
	const (
		scoreOffset       = 1.0 / 6
		leaderboardOffset = 1.0 / 4
		controlsOffset    = 3.0 / 4
	)

	width, height := s.Size()
	startX := width / 10

	// Draw title and game over message
	_ = renderer.DrawText(fmt.Sprintf("%s - %s", s.Config.Title, s.sceneName), startX, int(float64(height)*titleOffset), render.ColorWhite)

	if !s.nameEntered {
		// Draw name entry prompt
		_ = renderer.DrawText("Enter your name:", startX, int(float64(height)*scoreOffset), render.ColorWhite)
		if s.showOnBlink {
			_ = renderer.DrawText(s.name+"_", startX, int(float64(height)*scoreOffset)+2, render.ColorBrightMagenta)
		} else {
			_ = renderer.DrawText(s.name, startX, int(float64(height)*scoreOffset)+2, render.ColorBrightMagenta)
		}
	} else if s.showOnBlink {
		_ = renderer.DrawText(
			fmt.Sprintf("%d | %s > %s", s.Score, s.name, s.GetDetails()),
			startX,
			int(float64(height)*scoreOffset),
			render.ColorMagenta,
		)
	}

	// Draw leaderboard
	leaderboardY := int(float64(height) * leaderboardOffset)
	_ = renderer.DrawText("Top Scores:", startX, leaderboardY, render.ColorBlue)
	topScores := s.Leaderboard.TopScores(5)
	for i, entry := range topScores {
		_ = renderer.DrawText(fmt.Sprintf("%d | %s > %s", entry.Score, entry.Name, entry.Details), startX, leaderboardY+(i+1)*lineSpacing, render.ColorWhite)
	}

	// Draw controls
	controlsY := int(float64(height) * controlsOffset)
	_ = renderer.DrawText("Controls:", startX, controlsY, render.ColorBlue)
	_ = renderer.DrawText("Press Q to quit the game", startX, controlsY+lineSpacing, render.ColorWhite)
	_ = renderer.DrawText("Press ENTER to return to save/return to main menu", startX, controlsY+2*lineSpacing, render.ColorWhite)
}

func (s *GameOverScene) HandleInput(input core.InputEvent) error {
	if !s.nameEntered {
		switch input.Rune {
		case core.KeyBackspace:
			if len(s.name) > 0 {
				s.name = s.name[:len(s.name)-1]
			}
		case core.KeyEnter:
			if len(s.name) > 0 {
				s.nameEntered = true
				s.Logger.Info("Adding leaderboard entry...", "name", s.name, "score", s.Score)
				s.Leaderboard.Add(s.name, s.Score, s.GetDetails())
				err := s.Leaderboard.Save(s.Config.BoardFile)
				if err != nil {
					return err
				}
			}
		default:
			// Only allow printable characters
			if input.Rune >= 32 && input.Rune <= 126 && len(s.name) < 20 {
				s.name += string(input.Rune)
			}
		}
		return nil
	}

	switch input.Rune {
	case core.KeyEnter:
		s.Scenes.ChangeScene(MainMenuSceneID)
	case 'q', 'Q':
		return core.ErrQuitGame
	}
	return nil
}
