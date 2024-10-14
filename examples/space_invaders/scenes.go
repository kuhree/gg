package space_invaders

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/render"
	"github.com/kuhree/gg/internal/engine/scenes"
)

// BaseScene provides common functionality for all scenes
type BaseScene struct {
	game *Game
	name string
}

// Enter logs when a scene is entered
func (s *BaseScene) Enter() {
	s.game.logger.Info("Entering scene", "scene", s.name)
}

// Exit logs when a scene is exited
func (s *BaseScene) Exit() {
	s.game.logger.Info("Exiting scene", "scene", s.name)
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
	centerX := s.game.width / 2
	renderer.DrawText("SPACE INVADERS", centerX, s.game.height/4)
	if s.showPressEnter {
		renderer.DrawText("Press ENTER to start", centerX, s.game.height/2)
	}
	renderer.DrawText("Controls:", centerX, 2*s.game.height/3)
	renderer.DrawText("Arrow keys / WASD to move", centerX, 2*s.game.height/3+2)
	renderer.DrawText("SPACE to shoot", centerX, 2*s.game.height/3+4)
}

func (s *MainMenuScene) HandleInput(input core.InputEvent) error {
	if input.Key == core.KeyEnter {
		s.game.StartNewGame()
		s.game.sceneManager.ChangeScene(scenes.PlayingSceneID)
	}
	return nil
}

// PlayingScene methods

func (s *PlayingScene) Update(dt float64) {
	// Game logic updates are handled in the Game struct
	s.game.logger.Debug("Updating playing state")
	s.moveAliens(dt)
	s.moveBullets(dt)
	s.handleCollisions()
	s.checkGameOver()
}

// moveAliens updates the positions of all aliens
func (s *PlayingScene) moveAliens(dt float64) {
	moveDown := false
	alienWidth := float64(AlienSize)

	// Check if any alien has reached the screen edges
	for _, alien := range s.game.aliens {
		if (alien.Speed > 0 && alien.Position.X+alienWidth/2 >= float64(s.game.width)) ||
			(alien.Speed < 0 && alien.Position.X-alienWidth/2 <= 0) {
			moveDown = true
			break
		}
	}

	if moveDown {
		// Reverse direction and move down
		for _, alien := range s.game.aliens {
			alien.Speed = -alien.Speed
			alien.Position.Y += alienWidth / 2 // Move down by half the alien width
		}
		s.game.logger.Debug("Aliens moving down and reversing direction")
	} else {
		// Move horizontally
		for _, alien := range s.game.aliens {
			alien.Position.X += alien.Speed * dt
		}
	}

	// Increase speed slightly each time aliens move down
	if moveDown {
		for _, alien := range s.game.aliens {
			if alien.Speed > 0 {
				alien.Speed += 1
			} else {
				alien.Speed -= 1
			}
		}
	}
}

// moveBullets updates the positions of all bullets
func (s *PlayingScene) moveBullets(dt float64) {
	for i := len(s.game.bullets) - 1; i >= 0; i-- {
		bullet := s.game.bullets[i]
		bullet.Position.Y -= bullet.Speed * dt // Bullets move upwards

		// Remove bullets that are off-screen
		if bullet.Position.Y < 0 {
			s.game.bullets = append(s.game.bullets[:i], s.game.bullets[i+1:]...)
		}
	}
}

// movePlayer updates the player's position based on the given direction
func (s *PlayingScene) movePlayer(dx, dy int) {
	newX := s.game.player.Position.X + float64(dx)*s.game.player.Speed
	newY := s.game.player.Position.Y + float64(dy)*s.game.player.Speed

	// Clamp the player's position to stay within the game boundaries
	newX = clamp(newX, float64(PlayerSize)/2, float64(s.game.width)-float64(PlayerSize)/2)
	newY = clamp(newY, float64(PlayerSize)/2, float64(s.game.height)-float64(PlayerSize)/2)

	s.game.player.Position.X = newX
	s.game.player.Position.Y = newY

	s.game.logger.Debug("Player moved",
		"newX", newX,
		"newY", newY,
		"dx", dx,
		"dy", dy)
}

// LoadLevels loads level data from a JSON file
func (g *Game) LoadLevels(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open levels file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&g.levels); err != nil {
		return fmt.Errorf("failed to decode levels data: %w", err)
	}

	g.logger.Info("Levels loaded successfully", "count", len(g.levels))
	return nil
}

// fireBullet creates a new bullet from the player's position
func (s *PlayingScene) fireBullet() {
	bullet := &Bullet{
		GameObject: GameObject{
			Position: Vector2D{X: s.game.player.Position.X, Y: s.game.player.Position.Y - float64(PlayerSize)/2},
			Speed:    s.game.bulletSpeed,
		},
		Damage: 1,
	}
	s.game.bullets = append(s.game.bullets, bullet)
	s.game.logger.Info("Player fired a bullet", "bulletSpeed", s.game.bulletSpeed)
}

// handleCollisions detects and handles collisions between game objects
func (s *PlayingScene) handleCollisions() {
	// Check bullet collisions
	for i := len(s.game.bullets) - 1; i >= 0; i-- {
		bullet := s.game.bullets[i]
		for j, alien := range s.game.aliens {
			if checkCollision(bullet.Position, alien.Position, BulletSize, AlienSize) {
				s.game.score += alien.Points
				s.game.logger.Info("Score updated", "score", s.game.score)
				s.game.aliens = append(s.game.aliens[:j], s.game.aliens[j+1:]...)
				s.game.bullets = append(s.game.bullets[:i], s.game.bullets[i+1:]...)
				break
			}
		}

		for j, barrier := range s.game.barriers {
			if checkCollision(bullet.Position, barrier.Position, BulletSize, AlienSize) {
				s.game.barriers[j].Health -= bullet.Damage
				s.game.bullets = append(s.game.bullets[:i], s.game.bullets[i+1:]...)
				if s.game.barriers[j].Health <= 0 {
					s.game.barriers = append(s.game.barriers[:i], s.game.barriers[i+1:]...)
				}
				break
			}
		}
	}
}

// checkGameOver determines if the game should end
func (s *PlayingScene) checkGameOver() {
	if len(s.game.aliens) == 0 {
		s.game.currentLevel++
		s.game.logger.Info("Level completed", "newLevel", s.game.currentLevel+1)

		if s.game.currentLevel >= len(s.game.levels) {
			s.game.logger.Info("All levels completed, game won!")
			s.game.sceneManager.ChangeScene(scenes.GameOverSceneID)
			return
		}

		s.game.initializeLevel()
		return
	}

	for _, alien := range s.game.aliens {
		if alien.Position.Y+float64(AlienSize)/2 >= s.game.player.Position.Y-float64(PlayerSize)/2 {
			s.game.sceneManager.ChangeScene(scenes.GameOverSceneID)
			s.game.logger.Info("Game over: Aliens reached the bottom")
			return
		}
	}

	if s.game.player.lives <= 0 {
		s.game.sceneManager.ChangeScene(scenes.GameOverSceneID)
		s.game.logger.Info("Game over: Player out of lives")
	}
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
	s.game.renderer.DrawText(fmt.Sprintf("Score: %d", s.game.score), 1, 1)
	s.game.renderer.DrawText(fmt.Sprintf("Level: %d", s.game.currentLevel+1), 1, 2)
	s.game.renderer.DrawText(fmt.Sprintf("Enemies: %d", len(s.game.aliens)), 1, 3)
	s.game.renderer.DrawText(fmt.Sprintf("Lives: %d", s.game.player.lives), s.game.width-10, 1)
}

func (s *PlayingScene) HandleInput(input core.InputEvent) error {
	switch input.Key {
	case core.KeySpace:
		s.fireBullet()
	case core.KeyLeft:
		s.movePlayer(-1, 0)
	case core.KeyRight:
		s.movePlayer(1, 0)
	case core.KeyUp:
		s.movePlayer(0, -1)
	case core.KeyDown:
		s.movePlayer(0, 1)
	case core.KeyEscape:
		s.game.sceneManager.ChangeScene(scenes.PauseMenuSceneID)
	default:
		switch input.Rune {
		case 'w', 'W':
			s.movePlayer(0, -1)
		case 'a', 'A':
			s.movePlayer(-1, 0)
		case 's', 'S':
			s.movePlayer(0, 1)
		case 'd', 'D':
			s.movePlayer(1, 0)
		case ' ':
		s.fireBullet()
		}
	}
	return nil
}

// GameOverScene methods

func (s *GameOverScene) Draw(renderer *render.Renderer) {
	centerX := s.game.width / 2
	renderer.DrawText("GAME OVER", centerX, s.game.height/4)
	renderer.DrawText(fmt.Sprintf("Final Score: %d", s.game.score), centerX, s.game.height/3)
	renderer.DrawText(fmt.Sprintf("Final Level: %d", s.game.currentLevel+1), centerX, s.game.height/2)
	renderer.DrawText(fmt.Sprintf("Enemies Remaining: %d", len(s.game.aliens)), centerX, s.game.height/2+2)
	renderer.DrawText("Press ENTER to return to main menu", centerX, 2*s.game.height/3)
	renderer.DrawText("Press Q to quit the game", centerX, 2*s.game.height/3+2)
}

func (s *GameOverScene) HandleInput(input core.InputEvent) error {
	if input.Key == core.KeyEnter {
		s.game.sceneManager.ChangeScene(scenes.MainMenuSceneID)
		return nil
	}

	switch input.Rune {
	case 'q', 'Q':
		return core.ErrQuitGame
	}
	return nil
}

// PauseMenuScene methods

func (s *PauseMenuScene) Draw(renderer *render.Renderer) {
	centerX := s.game.width / 2
	renderer.DrawText("PAUSED", centerX, s.game.height/4)
	renderer.DrawText(fmt.Sprintf("Current Score: %d", s.game.score), centerX, s.game.height/3)
	renderer.DrawText(fmt.Sprintf("Current Level: %d", s.game.currentLevel+1), centerX, s.game.height/3+2)
	renderer.DrawText("Press ESC to resume", centerX, s.game.height/2)
	renderer.DrawText("Press R to restart level", centerX, s.game.height/2+2)
	renderer.DrawText("Press Q to quit", centerX, 2*s.game.height/3)
}

func (s *PauseMenuScene) HandleInput(input core.InputEvent) error {
	switch input.Key {
	case core.KeyEscape:
		s.game.sceneManager.ChangeScene(scenes.PlayingSceneID)
	default:
		switch input.Rune {
		case core.KeyQ:
			return core.ErrQuitGame
		}
	}
	return nil
}
