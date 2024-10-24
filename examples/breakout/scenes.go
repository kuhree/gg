package breakout

import (
	"fmt"
	"math"

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

	lives  int
	paddle *Paddle
	ball   *Ball
	bricks []*Brick
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
		lives: game.Config.InitialLives,
	}

	// Initialize paddle
	scene.paddle = &Paddle{
		GameObject: GameObject{
			Position: Vector2D{
				X: float64(game.Width) / 2,
				Y: float64(game.Height) - 2,
			},
			Width:  game.Config.PaddleWidth,
			Height: game.Config.PaddleHeight,
		},
		Speed: game.Config.PaddleSpeed,
	}

	// Initialize ball
	scene.ball = &Ball{
		GameObject: GameObject{
			Position: Vector2D{
				X: scene.paddle.Position.X + scene.paddle.Width/2,
				Y: scene.paddle.Position.Y - game.Config.BallSize,
			},
			Width:  game.Config.BallSize,
			Height: game.Config.BallSize,
		},
		Velocity: Vector2D{X: 0, Y: 0},
		Attached: true,
	}

	// Initialize bricks
	scene.initializeBricks()

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

	// Update paddle position
	if s.paddle.Position.X < 0 {
		s.paddle.Position.X = 0
	} else if s.paddle.Position.X > float64(s.Width)-s.paddle.Width {
		s.paddle.Position.X = float64(s.Width) - s.paddle.Width
	}

	// Update ball position
	if s.ball.Attached {
		s.ball.Position.X = s.paddle.Position.X + s.paddle.Width/2
		s.ball.Position.Y = s.paddle.Position.Y - 1
	} else {
		s.ball.Position.X += s.ball.Velocity.X * dt
		s.ball.Position.Y += s.ball.Velocity.Y * dt
	}

	s.updateCollisions(dt)
	ended, reason := s.checkGameState(dt)
	if ended {
		s.endGame(reason)
	}
}

func (s *PlayingScene) Draw(renderer *render.Renderer) {
	// Draw paddle
	for x := int(s.paddle.Position.X); x < int(s.paddle.Position.X+s.paddle.Width); x++ {
		y := int(s.paddle.Position.Y)
		_ = renderer.DrawChar('=', x, y, render.ColorCyan)
		s.drawObjOverlay(x, y, render.ColorCyan)
	}

	// Draw ball
	_ = renderer.DrawChar('O', int(s.ball.Position.X), int(s.ball.Position.Y), render.ColorWhite)
	s.drawObjOverlay(int(s.ball.Position.X), int(s.ball.Position.X), render.ColorWhite)

	// Draw bricks
	for _, brick := range s.bricks {
		for x := int(brick.Position.X); x < int(brick.Position.X+brick.Width); x++ {
			y := int(brick.Position.Y)
			_ = renderer.DrawChar('#', x, y, brick.Color)
			s.drawObjOverlay(x, y, brick.Color)
		}
	}

	// Draw score, level, lives
	_ = renderer.DrawText(fmt.Sprintf("Score: %d", s.Score), 1, 1, render.ColorWhite)
	_ = renderer.DrawText(fmt.Sprintf("Level: %d", s.CurrentLevel), 1, 2, render.ColorWhite)
	_ = renderer.DrawText(fmt.Sprintf("Lives: %d", s.lives), s.Width-10, 1, render.ColorWhite)
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
	case 'a', 'A':
		s.paddle.Position.X -= s.paddle.Speed
	case 'd', 'D':
		s.paddle.Position.X += s.paddle.Speed
	case ' ': // Spacebar launches the ball
		if s.ball.Attached {
			s.ball.Attached = false
			s.ball.Velocity = Vector2D{X: s.Config.BallVelocityX, Y: s.Config.BallVelocityY}
		}
	}

	return nil
}

// PlayingScene helpers

// updateCollisions detects and handles collisions between game objects
func (s *PlayingScene) updateCollisions(_ float64) {
	if s.ball.Attached {
		return
	}

	// Ball/Wall collisions
	if s.ball.Position.X <= 0 || s.ball.Position.X >= float64(s.Width) {
		s.ball.Velocity.X = -s.ball.Velocity.X
		// Ensure minimum X velocity to prevent vertical lock
		if math.Abs(s.ball.Velocity.X) < s.Config.BallMinXVelocity {
			if s.ball.Velocity.X < 0 {
				s.ball.Velocity.X = -s.Config.BallMinXVelocity
			} else {
				s.ball.Velocity.X = s.Config.BallMinXVelocity
			}
		}
	}
	if s.ball.Position.Y <= 0 {
		s.ball.Velocity.Y = -s.ball.Velocity.Y
		// Ensure minimum Y velocity to prevent horizontal lock at top
		if math.Abs(s.ball.Velocity.Y) < s.Config.BallMinYVelocity {
			s.ball.Velocity.Y = s.Config.BallMinYVelocity
		}
	}
	if s.ball.Position.Y >= float64(s.Height) {
		s.lives--
		s.resetBall()
	}

	// Ball/Paddle collision
	if s.ball.Position.Y >= s.paddle.Position.Y-1 &&
		s.ball.Position.Y <= s.paddle.Position.Y &&
		s.ball.Position.X >= s.paddle.Position.X &&
		s.ball.Position.X <= s.paddle.Position.X+s.paddle.Width {
		// Calculate reflection angle based on where ball hits paddle
		hitPos := (s.ball.Position.X - s.paddle.Position.X) / s.paddle.Width
		angle := (hitPos - 0.5) * 2 // -1 to 1
		s.ball.Velocity.X = angle * s.Config.BallSpeed
		s.ball.Velocity.Y = -s.Config.BallSpeed

		// Ensure minimum velocities
		if math.Abs(s.ball.Velocity.X) < s.Config.BallMinXVelocity {
			if s.ball.Velocity.X < 0 {
				s.ball.Velocity.X = -s.Config.BallMinXVelocity
			} else {
				s.ball.Velocity.X = s.Config.BallMinXVelocity
			}
		}
		if math.Abs(s.ball.Velocity.Y) < s.Config.BallMinYVelocity {
			if s.ball.Velocity.Y < 0 {
				s.ball.Velocity.Y = -s.Config.BallMinYVelocity
			} else {
				s.ball.Velocity.Y = s.Config.BallMinYVelocity
			}
		}
	}

	// Ball/Brick collisions
	for i := len(s.bricks) - 1; i >= 0; i-- {
		brick := s.bricks[i]
		if s.ball.Position.Y >= brick.Position.Y &&
			s.ball.Position.Y <= brick.Position.Y+brick.Height &&
			s.ball.Position.X >= brick.Position.X &&
			s.ball.Position.X <= brick.Position.X+brick.Width {
			brick.Health--
			if brick.Health <= 0 {
				s.Score += brick.Points
				// Remove brick
				s.bricks = append(s.bricks[:i], s.bricks[i+1:]...)
			}
			s.ball.Velocity.Y = -s.ball.Velocity.Y
			break
		}
	}
}

func (s *PlayingScene) resetBall() {
	s.ball.Attached = true
	s.ball.Position.X = s.paddle.Position.X + s.paddle.Width/2
	s.ball.Position.Y = s.paddle.Position.Y - s.Config.BallSize
	s.ball.Velocity = Vector2D{X: 0, Y: 0}
}

func (s *PlayingScene) initializeBricks() {
	s.bricks = make([]*Brick, 0)

	brickColors := []render.Color{
		render.ColorRed,
		render.ColorYellow,
		render.ColorGreen,
		render.ColorBlue,
	}

	brickWidth := s.Config.BrickWidth
	brickHeight := s.Config.BrickHeight
	rows := s.Config.BrickRows

	for row := 0; row < rows; row++ {
		y := s.Config.BrickStartY + float64(row)*s.Config.BrickSpacing
		bricksInRow := int(float64(s.Width) / brickWidth)

		for col := 0; col < bricksInRow; col++ {
			x := float64(col) * brickWidth

			brick := &Brick{
				GameObject: GameObject{
					Position: Vector2D{X: x, Y: y},
					Width:    brickWidth - 1,
					Height:   brickHeight,
				},
				Health: 1,
				Points: (rows - row) * 10,
				Color:  brickColors[row%len(brickColors)],
			}
			s.bricks = append(s.bricks, brick)
		}
	}
}

// checkGameState determines if the game should end
func (s *PlayingScene) checkGameState(_ float64) (bool, string) {
	if s.lives <= 0 {
		return true, "Out of lives"
	}

	if len(s.bricks) == 0 {
		s.CurrentLevel++
		s.initializeBricks()
		s.resetBall()
		return false, ""
	}

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
		debugInfo := []string{
			fmt.Sprintf("Pos: (%.1f,%.1f)", float64(x), float64(y)),
			fmt.Sprintf("Col: %d", color),
		}

		for i, info := range debugInfo {
			_ = s.Renderer.DrawText(info, x+1, y+i, color)
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
