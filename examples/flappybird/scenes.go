package flappybird

import (
	"fmt"
	"math/rand/v2"

	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/leaderboard"
	"github.com/kuhree/gg/internal/engine/render"
)

const (
	titleOffset = 1.0 / 10
	lineSpacing = 2

	// Difficulty bounds
	minPipeSpeed   = 10.0 // Minimum pipe speed
	maxPipeSpeed   = 50.0 // Maximum pipe speed
	minPipeGap     = 5.0  // Minimum gap between pipes
	maxPipeGap     = 20.0 // Maximum gap between pipes
	minGravity     = 9.8  // Minimum gravity
	maxGravity     = 25.0 // Maximum gravity
	minPipeSpacing = 10.0 // Minimum spacing between pipe pairs
	maxPipeSpacing = 40.0 // Maximum spacing between pipe pairs
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

	lives int
	bird  *Bird
	pipes []*Pipe

	pipeTimer   float64
	gameStarted bool

	// Current difficulty settings
	currentPipeSpeed   float64
	currentPipeGap     float64
	currentGravity     float64
	currentPipeSpacing float64
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
		lives:              game.Config.InitialLives,
		pipes:              make([]*Pipe, 0),
		currentPipeSpeed:   game.Config.PipeSpeed,
		currentPipeGap:     game.Config.PipeGap,
		currentGravity:     game.Config.BirdGravity,
		currentPipeSpacing: game.Config.PipeSpacing,
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

	if !s.gameStarted {
		// Initialize bird in center when game starts
		if s.bird == nil {
			s.bird = NewBird(float64(s.Width)/3, float64(s.Height)/2, s.Config)
			s.bird.Gravity = s.currentGravity
		}
		return
	}

	// Update bird physics
	s.bird.Velocity += s.bird.Gravity * dt
	s.bird.Position.Y += s.bird.Velocity * dt

	// Update pipes
	s.pipeTimer += dt
	if s.pipeTimer >= s.currentPipeSpacing/s.currentPipeSpeed {
		s.pipeTimer = 0
		s.spawnPipes()
	}

	for _, pipe := range s.pipes {
		pipe.Position.X -= s.currentPipeSpeed * dt
	}

	// Remove off-screen pipes
	newPipes := make([]*Pipe, 0)
	for _, pipe := range s.pipes {
		if pipe.Position.X > -pipe.Width {
			newPipes = append(newPipes, pipe)
		}
	}
	s.pipes = newPipes

	s.updateCollisions(dt)
	ended, reason := s.checkGameState(dt)
	if ended {
		s.endGame(reason)
	}
}

func (s *PlayingScene) spawnPipes() {
	gapY := float64(s.Height/2) + (rand.Float64()-0.5)*float64(s.Height/4)

	upperHeight := gapY - s.currentPipeGap/2
	lowerHeight := float64(s.Height) - (gapY + s.currentPipeGap/2)

	if upperHeight < s.Config.MinPipeHeight {
		upperHeight = s.Config.MinPipeHeight
	}
	if lowerHeight < s.Config.MinPipeHeight {
		lowerHeight = s.Config.MinPipeHeight
	}

	s.pipes = append(s.pipes,
		NewPipe(float64(s.Width), 0, upperHeight, s.Config.PipeWidth, true),
		NewPipe(float64(s.Width), gapY+s.Config.PipeGap/2, lowerHeight, s.Config.PipeWidth, false),
	)
}

func (s *PlayingScene) Draw(renderer *render.Renderer) {
	// Draw score, lives and debug info
	_ = renderer.DrawText(fmt.Sprintf("Score: %d", s.Score), 1, 1, render.ColorWhite)
	_ = renderer.DrawText(fmt.Sprintf("Lives: %d", s.lives), 1, 2, render.ColorWhite)
	_ = renderer.DrawText(fmt.Sprintf("Level: %d", s.CurrentLevel), 1, 3, render.ColorWhite)
	if s.Debug {
		_ = renderer.DrawText("Difficulty:", 1, 5, render.ColorBrightBlue)
		_ = renderer.DrawText(fmt.Sprintf("Speed: %.1f", s.currentPipeSpeed), 1, 6, render.ColorWhite)
		_ = renderer.DrawText(fmt.Sprintf("Gap: %.1f", s.currentPipeGap), 1, 7, render.ColorWhite)
		_ = renderer.DrawText(fmt.Sprintf("Spacing: %.1f", s.currentPipeSpacing), 1, 8, render.ColorWhite)
		_ = renderer.DrawText(fmt.Sprintf("Gravity: %.1f", s.currentGravity), 1, 9, render.ColorWhite)
	}

	// Draw start message
	if !s.gameStarted {
		msg := "Press SPACE to start!"
		x := (s.Width - len(msg)) / 2
		y := s.Height / 2
		_ = renderer.DrawText(msg, x, y, render.ColorBrightMagenta)
	}

	// Draw bird
	if s.bird != nil {
		_ = renderer.DrawChar(s.bird.Character, int(s.bird.Position.X), int(s.bird.Position.Y), s.bird.Color)
		s.drawObjOverlay(int(s.bird.Position.X), int(s.bird.Position.Y), render.ColorWhite)
	}

	// Draw pipes
	for _, pipe := range s.pipes {
		pipeX := int(pipe.Position.X)
		if pipe.IsUpperPipe {
			for y := 0; y < int(pipe.Height); y++ {
				_ = renderer.DrawChar('|', pipeX, y, pipe.Color)
				for i := 0; i < int(pipe.Width); i++ {
					_ = renderer.DrawChar('|', pipeX+i, y, pipe.Color)
				}
			}
		} else {
			startY := int(pipe.Position.Y)
			for y := startY; y < startY+int(pipe.Height); y++ {
				_ = renderer.DrawChar('|', pipeX, y, pipe.Color)
				for i := 0; i < int(pipe.Width); i++ {
					_ = renderer.DrawChar('|', pipeX+i, y, pipe.Color)
				}
			}
		}

		s.drawObjOverlay(int(pipeX), int(pipe.Position.Y), render.ColorWhite)
	}
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
	case '=', '+':
		s.increaseLevel()
	case '-', '_':
		s.decreaseLevel()
	case ' ':
		if !s.gameStarted {
			s.gameStarted = true
		}
		if s.bird != nil && !s.bird.IsDead {
			s.bird.Velocity = s.bird.JumpForce
		}
	}

	return nil
}

// PlayingScene helpers

// updateCollisions detects and handles collisions between game objects
func (s *PlayingScene) updateCollisions(_ float64) {
	if s.bird == nil || !s.gameStarted {
		return
	}

	// Check floor/ceiling collisions
	if s.bird.Position.Y < 0 || s.bird.Position.Y >= float64(s.Height) {
		s.bird.IsDead = true
		return
	}

	// Check pipe collisions
	birdX := int(s.bird.Position.X)
	birdY := int(s.bird.Position.Y)

	for _, pipe := range s.pipes {
		pipeX := int(pipe.Position.X)

		// Only check pipes the bird is passing through
		if birdX >= pipeX && birdX <= pipeX+int(pipe.Width) {
			if pipe.IsUpperPipe {
				if birdY < int(pipe.Height) {
					s.bird.IsDead = true
					return
				}
			} else {
				if birdY >= int(pipe.Position.Y) {
					s.bird.IsDead = true
					return
				}
			}
		}

		// Score point when passing pipe (only count lower pipe to avoid double scoring)
		if birdX == pipeX+int(pipe.Width) && !pipe.IsUpperPipe && !pipe.Scored {
			s.Score++
			pipe.Scored = true

			// Increase difficulty every 5 points
			if s.Score%5 == 0 {
				s.increaseLevel()
			}
		}
	}
}

// checkGameState determines if the game should end
func (s *PlayingScene) checkGameState(_ float64) (bool, string) {
	if s.bird != nil && s.bird.IsDead {
		s.lives--
		if s.lives <= 0 {
			return true, "Out of lives"
		}

		// Reset for next life
		s.bird = nil
		s.pipes = make([]*Pipe, 0)
		s.gameStarted = false
	}

	return false, ""
}

// increaseLevel increases the game difficulty by adjusting various parameters
func (s *PlayingScene) increaseLevel() {
	s.CurrentLevel++
	s.currentPipeSpeed = min(maxPipeSpeed, s.currentPipeSpeed*1.2)       // Increase speed by 20%
	s.currentPipeGap = max(minPipeGap, s.currentPipeGap*0.9)             // Decrease gap by 10%
	s.currentGravity = min(maxGravity, s.currentGravity*1.1)             // Increase gravity by 10%
	s.currentPipeSpacing = max(minPipeSpacing, s.currentPipeSpacing*0.9) // Decrease spacing by 10%
}

// decreaseLevel decreases the game difficulty if above level 0
func (s *PlayingScene) decreaseLevel() {
	if s.CurrentLevel > 0 {
		s.CurrentLevel--
		s.currentPipeSpeed = max(minPipeSpeed, s.currentPipeSpeed/1.2)       // Decrease speed by 20%
		s.currentPipeGap = min(maxPipeGap, s.currentPipeGap/0.9)             // Increase gap by 10%
		s.currentGravity = max(minGravity, s.currentGravity/1.1)             // Decrease gravity by 10%
		s.currentPipeSpacing = min(maxPipeSpacing, s.currentPipeSpacing/0.9) // Increase spacing by 10%
	}
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
		_ = s.Renderer.DrawChar(render.FullBlock, x, y, color)
	}

	if s.Debug {
		debugInfo := []string{
			("┌─ Position ─────────┐"),
			fmt.Sprintf("│ X: %-15.1f│", float64(x)),
			fmt.Sprintf("│ Y: %-15.1f│", float64(y)),
			("└───────────────────┘"),
		}

		// Add bird-specific debug info
		if s.bird != nil && x == int(s.bird.Position.X) && y == int(s.bird.Position.Y) {
			debugInfo = append(debugInfo,
				("┌─ Bird Stats ─────────┐"),
				fmt.Sprintf("│ Velocity: %-10.1f│", s.bird.Velocity),
				fmt.Sprintf("│ Gravity:  %-10.1f│", s.bird.Gravity),
				fmt.Sprintf("│ Jump:     %-10.1f│", s.bird.JumpForce),
				("└────────────────────┘"),
			)
		}

		// Add pipe-specific debug info
		for _, pipe := range s.pipes {
			if x == int(pipe.Position.X) && y == int(pipe.Position.Y) {
				pipeType := "Lower"
				if pipe.IsUpperPipe {
					pipeType = "Upper"
				}
				debugInfo = append(debugInfo,
					("┌─ Pipe Stats ─────────┐"),
					fmt.Sprintf("│ Type:   %-11s│", pipeType),
					fmt.Sprintf("│ Height: %-11.1f│", pipe.Height),
					fmt.Sprintf("│ Width:  %-11.1f│", pipe.Width),
					fmt.Sprintf("│ Scored: %-11v│", pipe.Scored),
					("└────────────────────┘"),
				)
			}
		}

		// Draw debug info offset to the right
		for i, info := range debugInfo {
			_ = s.Renderer.DrawText(info, x+3, y+i-len(debugInfo)/2, render.ColorBrightBlue)
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
		"%dW*%dH|L%d|S%d||PS%.1f|PW%.1f|PG%.1f|IL%d|GV%.1f|JF%.1f",
		width, height,
		s.CurrentLevel,
		s.Score,
		s.Config.PipeSpeed,
		s.Config.PipeWidth,
		s.Config.PipeGap,
		s.Config.InitialLives,
		s.Config.BirdGravity,
		s.Config.BirdJumpForce,
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

	if s.Score > 0 && !s.nameEntered {
		// Draw name entry prompt and score
		_ = renderer.DrawText(fmt.Sprintf("Score: %d", s.Score), startX, int(float64(height)*scoreOffset), render.ColorWhite)
		_ = renderer.DrawText("Enter your name to save score (or press Q to skip):", startX, int(float64(height)*scoreOffset)+1, render.ColorWhite)
		if s.showOnBlink {
			_ = renderer.DrawText(s.name+"_", startX, int(float64(height)*scoreOffset)+2, render.ColorBrightMagenta)
		} else {
			_ = renderer.DrawText(s.name, startX, int(float64(height)*scoreOffset)+2, render.ColorBrightMagenta)
		}
	} else if s.Score > 0 && s.showOnBlink {
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
	topScores := s.Leaderboard.TopScores(s.Config.LeaderboardSize)
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
	switch input.Rune {
	case 'q', 'Q':
		if !s.nameEntered && s.Score > 0 {
			s.Logger.Info("Skipping leaderboard entry")
			s.nameEntered = true
		}
		return core.ErrQuitGame
	case core.KeyEnter:
		if !s.nameEntered {
			if len(s.name) > 0 && s.Score > 0 {
				s.nameEntered = true
				s.Logger.Info("Adding leaderboard entry...", "name", s.name, "score", s.Score)
				s.Leaderboard.Add(s.name, s.Score, s.GetDetails())
				err := s.Leaderboard.Save(s.Config.BoardFile)
				if err != nil {
					return err
				}
			}
		} else {
			s.Scenes.ChangeScene(MainMenuSceneID)
		}
	case core.KeyBackspace:
		if !s.nameEntered && len(s.name) > 0 {
			s.name = s.name[:len(s.name)-1]
		}
	default:
		if !s.nameEntered {
			// Only allow printable characters
			if input.Rune >= 32 && input.Rune <= 126 && len(s.name) < s.Config.MaxNameLength {
				s.name += string(input.Rune)
			}
		}
	}
	return nil
}
