package space_invaders

import (
	"fmt"

	"github.com/kuhree/gg/examples/space_invaders/scenes"
	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/render"
)

// BaseScene provides common functionality for all scenes
type BaseScene struct {
	game *Game
	name string
}

// Enter logs when a scene is entered
func (s *BaseScene) Enter() {
	s.game.Logger().Info("Entering scene", "scene", s.name)
}

// Exit logs when a scene is exited
func (s *BaseScene) Exit() {
	s.game.Logger().Info("Exiting scene", "scene", s.name)
}

// Update is a no-op for scenes that don't need updates
func (s *BaseScene) Update(dt float64) {}

// HandleInput is a no-op for scenes that don't handle input
func (s *BaseScene) HandleInput(input core.InputEvent) error {
	return nil
}

// MainMenuScene represents the main menu
type MainMenuScene struct {
	BaseScene
	blinkTimer     float64
	blinkInterval  float64
	showPressEnter bool
}

// PlayingScene represents the main gameplay
type PlayingScene struct {
	BaseScene
}

// GameOverScene represents the game over screen
type GameOverScene struct {
	BaseScene
}

// PauseMenuScene represents the pause menu
type PauseMenuScene struct {
	BaseScene
}

// NewMainMenuScene creates a new main menu scene
func NewMainMenuScene(game *Game) *MainMenuScene {
	return &MainMenuScene{
		BaseScene: BaseScene{
			game: game,
			name: "Main Menu",
		},
		blinkInterval:  0.5,
		showPressEnter: true,
	}
}

// NewPlayingScene creates a new playing scene
func NewPlayingScene(game *Game) *PlayingScene {
	return &PlayingScene{
		BaseScene: BaseScene{
			game: game,
			name: "Playing",
		},
	}
}

// NewGameOverScene creates a new game over scene
func NewGameOverScene(game *Game) *GameOverScene {
	return &GameOverScene{
		BaseScene: BaseScene{
			game: game,
			name: "Game Over",
		},
	}
}

// NewPauseMenuScene creates a new pause menu scene
func NewPauseMenuScene(game *Game) *PauseMenuScene {
	return &PauseMenuScene{
		BaseScene: BaseScene{
			game: game,
			name: "Pause Menu",
		},
	}
}

// MainMenuScene methods

func (s *MainMenuScene) Update(dt float64) {
	s.blinkTimer += dt
	if s.blinkTimer >= s.blinkInterval {
		s.blinkTimer = 0
		s.showPressEnter = !s.showPressEnter
	}
}

func (s *MainMenuScene) Draw(renderer *render.Renderer) {
	renderer.DrawText("SPACE INVADERS", s.game.Width()/2, s.game.Height()/3)
	if s.showPressEnter {
		renderer.DrawText("Press ENTER to start", s.game.Width()/2, s.game.Height()/2)
	}
}

func (s *MainMenuScene) HandleInput(input core.InputEvent) error {
	if input.Key == core.KeyEnter {
		s.game.StartNewGame()
		s.game.ChangeScene(scenes.PlayingSceneID)
	}
	return nil
}

// PlayingScene methods

func (s *PlayingScene) Update(dt float64) {
	// Game logic updates are handled in the Game struct
	s.game.Logger().Debug("Updating playing state")
	s.game.moveAliens(dt)
	s.game.moveBullets(dt)
	s.game.handleCollisions()
	s.game.checkGameOver()
}

func (s *PlayingScene) Draw(renderer *render.Renderer) {
	// Draw player
	s.game.renderer.DrawRect(int(s.game.player.Position.X-float64(PlayerSize)/2), int(s.game.player.Position.Y-float64(PlayerSize)/2), PlayerSize, PlayerSize, 'P')

	// Draw aliens
	for _, alien := range s.game.aliens {
		s.game.renderer.DrawRect(int(alien.Position.X-float64(AlienSize)/2), int(alien.Position.Y-float64(AlienSize)/2), AlienSize, AlienSize, 'A')
	}

	// Draw bullets
	for _, bullet := range s.game.bullets {
		s.game.renderer.DrawRect(int(bullet.Position.X-float64(BulletSize)/2), int(bullet.Position.Y-float64(BulletSize)/2), BulletSize, BulletSize, '*')
	}

	// Draw barriers
	for _, barrier := range s.game.barriers {
		s.game.renderer.DrawRect(int(barrier.Position.X-float64(BarrierSize)/2), int(barrier.Position.Y-float64(BarrierSize)/2), BarrierSize, BarrierSize, '+')
	}

	// Draw score, level, remaining enemies, and lives
	s.game.renderer.DrawText(fmt.Sprintf("Score: %d | Level: %d | Enemies: %d", s.game.score, s.game.currentLevel+1, len(s.game.aliens)), 1, 1)
	s.game.renderer.DrawText(fmt.Sprintf("Lives: %d", s.game.player.lives), s.game.width-10, 1)
}

func (s *PlayingScene) HandleInput(input core.InputEvent) error {
	switch input.Key {
	case core.KeySpace:
		s.game.FireBullet()
	case core.KeyLeft:
		s.game.MovePlayer(-1, 0)
	case core.KeyRight:
		s.game.MovePlayer(1, 0)
	case core.KeyUp:
		s.game.MovePlayer(0, -1)
	case core.KeyDown:
		s.game.MovePlayer(0, 1)
	case core.KeyEscape:
		s.game.ChangeScene(scenes.PauseMenuSceneID)
	default:
		switch input.Rune {
		case 'w', 'W':
			s.game.MovePlayer(0, -1)
		case 'a', 'A':
			s.game.MovePlayer(-1, 0)
		case 's', 'S':
			s.game.MovePlayer(0, 1)
		case 'd', 'D':
			s.game.MovePlayer(1, 0)
		case ' ':
			s.game.FireBullet()
		}
	}
	return nil
}

// GameOverScene methods

func (s *GameOverScene) Draw(renderer *render.Renderer) {
	renderer.DrawText("GAME OVER", s.game.Width()/2, s.game.Height()/3)
	renderer.DrawText(fmt.Sprintf("Final Score: %d", s.game.GetScore()), s.game.Width()/2, s.game.Height()/2)
	renderer.DrawText(fmt.Sprintf("Final Level: %d", s.game.currentLevel), s.game.Width()/2, s.game.Height()/2)
	renderer.DrawText(fmt.Sprintf("Enemies Remaining: %d", len(s.game.aliens)), s.game.Width()/2, s.game.Height()/2)
	renderer.DrawText("Press ENTER to return to main menu", s.game.Width()/2, 2*s.game.Height()/3)
}

func (s *GameOverScene) HandleInput(input core.InputEvent) error {
	if input.Key == core.KeyEnter {
		s.game.ChangeScene(scenes.MainMenuSceneID)
	}
	return nil
}

// PauseMenuScene methods

func (s *PauseMenuScene) Draw(renderer *render.Renderer) {
	renderer.DrawText("PAUSED", s.game.Width()/2, s.game.Height()/3)
	renderer.DrawText("Press ESC to resume", s.game.Width()/2, s.game.Height()/2)
	renderer.DrawText("Press Q to quit", s.game.Width()/2, 2*s.game.Height()/3)
}

func (s *PauseMenuScene) HandleInput(input core.InputEvent) error {
	switch input.Key {
	case core.KeyEscape:
		s.game.ChangeScene(scenes.PlayingSceneID)
	default:
		switch input.Rune {
		case core.KeyQ:
			s.game.ChangeScene(scenes.MainMenuSceneID)
		}
	}
	return nil
}
