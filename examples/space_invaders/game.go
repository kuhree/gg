package space_invaders

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/render"
)

// GameMode represents the current state of the game
type GameMode int

const (
	MainMenu GameMode = iota
	Playing
	GameOver
	PauseMenu
)

// AlienType represents different types of aliens
type AlienType int

const (
	LowAlien AlienType = iota
	MediumAlien
	HighAlien
)

// Vector2D represents a 2D position or velocity
type Vector2D struct {
	X, Y float64
}

// GameObject represents a basic game entity with position and speed
type GameObject struct {
	Position Vector2D
	Speed    float64
}

// Player represents the player's ship
type Player struct {
	GameObject

	lives int
}

// Alien represents an enemy alien
type Alien struct {
	GameObject
	Type   AlienType
	Points int
}

// Bullet represents a projectile fired by the player or aliens
type Bullet struct {
	GameObject
	Damage int
}

// Barrier represents a defensive structure
type Barrier struct {
	GameObject
	Health int
}

// LevelConfig represents the configuration for a single level
type LevelConfig struct {
	AlienRows     int     `json:"alienRows"`
	AliensPerRow  int     `json:"aliensPerRow"`
	AlienSpeed    float64 `json:"alienSpeed"`
	BarrierCount  int     `json:"barrierCount"`
	BarrierHealth int     `json:"barrierHealth"`
}

// Game represents the Space Invaders game state and logic
type Game struct {
	renderer *render.Renderer
	logger   *slog.Logger
	currentScene Scene

	mainMenuScene  *MainMenuScene
	playingScene   *PlayingScene
	gameOverScene  *GameOverScene
	pauseMenuScene *PauseMenuScene

	score        int
	currentLevel int
	levels       []LevelConfig
	lastTime     time.Time

	player   *Player
	aliens   []*Alien
	bullets  []*Bullet
	barriers []*Barrier

	width  int
	height int

	// Main menu fields
	blinkInterval  float64
	showPressEnter bool

	bulletSpeed float64
}

const (
	PlayerSpeed = 1
	BulletSpeed = 60
	AlienSpeed  = 10
	PlayerSize  = 1
	AlienSize   = 2
	BulletSize  = 1
	BarrierSize = 3
)

// Scene represents a game scene (e.g., MainMenu, Playing, GameOver)
type Scene interface {
	Enter()
	Exit()
	Update(dt float64)
	Draw(renderer *render.Renderer)
	HandleInput(input core.InputEvent) error
}

// NewGame creates a new instance of the Space Invaders game
func NewGame(renderer *render.Renderer, logger *slog.Logger) *Game {
	width, height := renderer.Size()

	game := &Game{
		renderer: renderer,
		logger:   logger,
		lastTime: time.Now(),
		width:    width,
		height:   height,
		player: &Player{
			GameObject: GameObject{
				Position: Vector2D{X: float64(width) / 2, Y: float64(height) - 3},
				Speed:    PlayerSpeed,
			},
			lives: 3,
		},
		blinkInterval:  0.5,
		showPressEnter: true,
		bulletSpeed:    60, // Initial bullet speed
	}

	game.mainMenuScene = &MainMenuScene{game: game}
	game.playingScene = &PlayingScene{game: game}
	game.gameOverScene = &GameOverScene{game: game}
	game.pauseMenuScene = &PauseMenuScene{game: game}

	game.ChangeScene(game.mainMenuScene)

	if err := game.LoadLevels("examples/space_invaders/levels.json"); err != nil {
		logger.Error("Failed to load levels", "error", err)
	} else {
		logger.Info("Levels loaded successfully", "count", len(game.levels))
		if len(game.levels) == 1 {
			logger.Info("Only one level loaded. Game will end after completing this level.")
		}
	}

	return game
}

// Init initializes the game
func (g *Game) Init() error {
	g.logger.Info("Space Invaders game initialized")
	return nil
}

// Cleanup performs any necessary cleanup
func (g *Game) Cleanup() {
	g.currentScene = g.gameOverScene
	g.logger.Info("Space Invaders game cleaned up")
}

// Draw renders the game state
func (g *Game) Draw() {
	g.renderer.Clear()
	g.currentScene.Draw(g.renderer)
	g.renderer.Render()
}

// Update updates the game state
func (g *Game) Update(dt float64) error {
	g.currentScene.Update(dt)
	return nil
}

// HandleInput processes user input
func (g *Game) HandleInput(input core.InputEvent) error {
	return g.currentScene.HandleInput(input)
}

// moveAliens updates the positions of all aliens
func (g *Game) moveAliens(dt float64) {
	moveDown := false
	alienWidth := float64(AlienSize)

	// Check if any alien has reached the screen edges
	for _, alien := range g.aliens {
		if (alien.Speed > 0 && alien.Position.X+alienWidth/2 >= float64(g.width)) ||
			(alien.Speed < 0 && alien.Position.X-alienWidth/2 <= 0) {
			moveDown = true
			break
		}
	}

	if moveDown {
		// Reverse direction and move down
		for _, alien := range g.aliens {
			alien.Speed = -alien.Speed
			alien.Position.Y += alienWidth / 2 // Move down by half the alien width
		}
		g.logger.Debug("Aliens moving down and reversing direction")
	} else {
		// Move horizontally
		for _, alien := range g.aliens {
			alien.Position.X += alien.Speed * dt
		}
	}

	// Increase speed slightly each time aliens move down
	if moveDown {
		for _, alien := range g.aliens {
			if alien.Speed > 0 {
				alien.Speed += 1
			} else {
				alien.Speed -= 1
			}
		}
	}
}

// moveBullets updates the positions of all bullets
func (g *Game) moveBullets(dt float64) {
	for i := len(g.bullets) - 1; i >= 0; i-- {
		bullet := g.bullets[i]
		bullet.Position.Y -= bullet.Speed * dt // Bullets move upwards

		// Remove bullets that are off-screen
		if bullet.Position.Y < 0 {
			g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
		}
	}
}

// handleCollisions detects and handles collisions between game objects
func (g *Game) handleCollisions() {
	// Check bullet collisions
	for i := len(g.bullets) - 1; i >= 0; i-- {
		bullet := g.bullets[i]
		for j, alien := range g.aliens {
			if checkCollision(bullet.Position, alien.Position, BulletSize, AlienSize) {
				g.updateScore(alien.Points)
				g.aliens = append(g.aliens[:j], g.aliens[j+1:]...)
				g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
				break
			}
		}

		for j, barrier := range g.barriers {
			if checkCollision(bullet.Position, barrier.Position, BulletSize, AlienSize) {
				g.barriers[j].Health -= bullet.Damage
				g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
				if g.barriers[j].Health <= 0 {
					g.barriers = append(g.barriers[:i], g.barriers[i+1:]...)
				}
				break
			}
		}
	}
}

// checkGameOver determines if the game should end
func (g *Game) checkGameOver() {
	if len(g.aliens) == 0 {
		g.currentLevel++
		g.logger.Info("Level completed", "newLevel", g.currentLevel+1)

		if g.currentLevel >= len(g.levels) {
			g.logger.Info("All levels completed, game won!")
			g.ChangeScene(g.gameOverScene)
			return
		}

		g.initializeLevel()
		return
	}

	for _, alien := range g.aliens {
		if alien.Position.Y+float64(AlienSize)/2 >= g.player.Position.Y-float64(PlayerSize)/2 {
			g.ChangeScene(g.gameOverScene)
			g.logger.Info("Game over: Aliens reached the bottom")
			return
		}
	}

	if g.player.lives <= 0 {
		g.ChangeScene(g.gameOverScene)
		g.logger.Info("Game over: Player out of lives")
	}
}

// Helper functions

func checkCollision(pos1, pos2 Vector2D, size1, size2 float64) bool {
	return abs(pos1.X-pos2.X) < (size1+size2)/2 && abs(pos1.Y-pos2.Y) < (size1+size2)/2
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// fireBullet creates a new bullet from the player's position
func (g *Game) fireBullet() {
	bullet := &Bullet{
		GameObject: GameObject{
			Position: Vector2D{X: g.player.Position.X, Y: g.player.Position.Y - float64(PlayerSize)/2},
			Speed:    g.bulletSpeed,
		},
		Damage: 1,
	}
	g.bullets = append(g.bullets, bullet)
	g.logger.Info("Player fired a bullet", "bulletSpeed", g.bulletSpeed)
}

// updateScore increases the player's score
func (g *Game) updateScore(points int) {
	g.score += points
	g.logger.Info("Score updated", "score", g.score)
}

// startNewGame initializes a new game
func (g *Game) startNewGame() {
	g.logger.Info("Starting new game")
	g.score = 0
	g.currentLevel = 0
	g.player.lives = 3
	g.bulletSpeed = BulletSpeed
	g.initializeLevel()
}

// movePlayer updates the player's position based on the given direction
func (g *Game) movePlayer(dx, dy int) {
	newX := g.player.Position.X + float64(dx)*g.player.Speed
	newY := g.player.Position.Y + float64(dy)*g.player.Speed

	// Clamp the player's position to stay within the game boundaries
	newX = clamp(newX, float64(PlayerSize)/2, float64(g.width)-float64(PlayerSize)/2)
	newY = clamp(newY, float64(PlayerSize)/2, float64(g.height)-float64(PlayerSize)/2)

	g.player.Position.X = newX
	g.player.Position.Y = newY

	g.logger.Debug("Player moved",
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

// initializeLevel configures the game state for the current level
func (g *Game) initializeLevel() {
	if g.currentLevel >= len(g.levels) {
		g.logger.Info("All levels completed, restarting from the first level")
		g.currentLevel = 0
	}

	if len(g.levels) == 0 {
		g.logger.Error("No levels loaded")
		return
	}

	levelData := g.levels[g.currentLevel]

	// Reset game entities
	g.aliens = nil
	g.bullets = nil
	g.barriers = nil

	// Setup aliens
	alienWidth := (float64(g.width) - 4.0 - float64(levelData.AliensPerRow-1)*2.0) / float64(levelData.AliensPerRow)
	alienHeight := 1.0

	for row := 0; row < levelData.AlienRows; row++ {
		for col := 0; col < levelData.AliensPerRow; col++ {
			alienType := AlienType(row / 2)
			alien := &Alien{
				GameObject: GameObject{
					Position: Vector2D{
						X: 2.0 + float64(col)*(alienWidth+2.0) + alienWidth/2,
						Y: 2.0 + float64(row)*(alienHeight+2.0) + alienHeight/2,
					},
					Speed: levelData.AlienSpeed,
				},
				Type:   alienType,
				Points: (3 - int(alienType)) * 10,
			}
			g.aliens = append(g.aliens, alien)
		}
	}

	// Setup barriers
	for i := 0; i < levelData.BarrierCount; i++ {
		barrier := &Barrier{
			GameObject: GameObject{
				Position: Vector2D{
					X: float64(i+1)*(float64(g.width)/(float64(levelData.BarrierCount)+1)) - float64(BarrierSize)/2,
					Y: float64(g.height) - 5,
				},
			},
			Health: levelData.BarrierHealth,
		}
		g.barriers = append(g.barriers, barrier)
	}

	// Set player position
	g.player.Position = Vector2D{X: float64(g.width) / 2, Y: float64(g.height) - 3}

	g.logger.Info("Level setup complete",
		"level", g.currentLevel+1,
		"aliens", len(g.aliens),
		"barriers", len(g.barriers),
		"alienSpeed", levelData.AlienSpeed,
		"bulletSpeed", g.bulletSpeed)
}

type MainMenuScene struct {
	game           *Game
	blinkTimer     float64
	blinkInterval  float64
	showPressEnter bool
}

type PlayingScene struct {
	game *Game
}

type GameOverScene struct {
	game *Game
}

type PauseMenuScene struct {
	game            *Game
}

func (s *MainMenuScene) Enter() {
	s.game.logger.Info("Entering Main Menu")
	s.blinkInterval = 0.5
	s.showPressEnter = true
}

func (s *MainMenuScene) Exit() {
	s.game.logger.Info("Exiting Main Menu")
}

func (s *MainMenuScene) Update(dt float64) {
	s.blinkTimer += dt
	if s.blinkTimer >= s.blinkInterval {
		s.blinkTimer = 0
		s.showPressEnter = !s.showPressEnter
	}
}

func (s *MainMenuScene) Draw(renderer *render.Renderer) {
	renderer.DrawText("SPACE INVADERS", s.game.width/2, s.game.height/3)
	if s.showPressEnter {
		renderer.DrawText("Press ENTER to start", s.game.width/2, s.game.height/2)
	}
}

func (s *MainMenuScene) HandleInput(input core.InputEvent) error {
	if input.Key == core.KeyEnter {
		s.game.ChangeScene(s.game.playingScene)
	}
	return nil
}

func (g *Game) ChangeScene(newScene Scene) {
	if g.currentScene != nil {
		g.currentScene.Exit()
	}
	g.currentScene = newScene
	g.currentScene.Enter()
}

func (s *PlayingScene) Enter() {
	s.game.logger.Info("Entering Playing Scene")
	s.game.startNewGame()
}

func (s *PlayingScene) Exit() {
	s.game.logger.Info("Exiting Playing Scene")
}

func (s *PlayingScene) Update(dt float64) {
	s.game.moveAliens(dt)
	s.game.moveBullets(dt)
	s.game.handleCollisions()
	s.game.checkGameOver()
}

func (s *PlayingScene) Draw(renderer *render.Renderer) {
	// Draw player
	renderer.DrawRect(int(s.game.player.Position.X-float64(PlayerSize)/2), int(s.game.player.Position.Y-float64(PlayerSize)/2), PlayerSize, PlayerSize, 'P')

	// Draw aliens
	for _, alien := range s.game.aliens {
		renderer.DrawRect(int(alien.Position.X-float64(AlienSize)/2), int(alien.Position.Y-float64(AlienSize)/2), AlienSize, AlienSize, 'A')
	}

	// Draw bullets
	for _, bullet := range s.game.bullets {
		renderer.DrawRect(int(bullet.Position.X-float64(BulletSize)/2), int(bullet.Position.Y-float64(BulletSize)/2), BulletSize, BulletSize, '*')
	}

	// Draw barriers
	for _, barrier := range s.game.barriers {
		renderer.DrawRect(int(barrier.Position.X-float64(BarrierSize)/2), int(barrier.Position.Y-float64(BarrierSize)/2), BarrierSize, BarrierSize, '+')
	}

	// Draw score, level, remaining enemies, and lives
	renderer.DrawText(fmt.Sprintf("Score: %d | Level: %d | Enemies: %d", s.game.score, s.game.currentLevel+1, len(s.game.aliens)), 1, 1)
	renderer.DrawText(fmt.Sprintf("Lives: %d", s.game.player.lives), s.game.width-10, 1)
}

func (s *PlayingScene) HandleInput(input core.InputEvent) error {
	switch input.Key {
	case core.KeySpace:
		s.game.fireBullet()
	case core.KeyBackspace:
		s.game.ChangeScene(s.game.pauseMenuScene)
	case core.KeyLeft:
		s.game.movePlayer(-1, 0)
	case core.KeyRight:
		s.game.movePlayer(1, 0)
	case core.KeyUp:
		s.game.movePlayer(0, -1)
	case core.KeyDown:
		s.game.movePlayer(0, 1)
	default:
		switch input.Rune {
		case 'w', 'W':
			s.game.movePlayer(0, -1)
		case 'a', 'A':
			s.game.movePlayer(-1, 0)
		case 's', 'S':
			s.game.movePlayer(0, 1)
		case 'd', 'D':
			s.game.movePlayer(1, 0)
		case ' ':
			s.game.fireBullet()
		}
	}
	return nil
}

func (s *GameOverScene) Enter() {
	s.game.logger.Info("Entering Game Over Scene")
}

func (s *GameOverScene) Exit() {
	s.game.logger.Info("Exiting Game Over Scene")
}

func (s *GameOverScene) Update(dt float64) {
	// No update logic needed for game over scene
}

func (s *GameOverScene) Draw(renderer *render.Renderer) {
	renderer.DrawText("GAME OVER", s.game.width/2, s.game.height/3)
	renderer.DrawText(fmt.Sprintf("Final Score: %d", s.game.score), s.game.width/2, s.game.height/2)
	renderer.DrawText("Press ENTER to restart", s.game.width/2, 2*s.game.height/3)
}

func (s *GameOverScene) HandleInput(input core.InputEvent) error {
	if input.Key == core.KeyEnter {
		s.game.ChangeScene(s.game.mainMenuScene)
	}
	return nil
}

func (s *PauseMenuScene) Enter() {
	s.game.logger.Info("Entering Pause Menu Scene")
}

func (s *PauseMenuScene) Exit() {
	s.game.logger.Info("Exiting Pause Menu Scene")
}

func (s *PauseMenuScene) Update(dt float64) {
	// No update logic needed for pause menu scene
}

func (s *PauseMenuScene) Draw(renderer *render.Renderer) {
	renderer.DrawText("PAUSED", s.game.width/2, s.game.height/3)
	renderer.DrawText("Press P to resume", s.game.width/2, s.game.height/2)
}

func (s *PauseMenuScene) HandleInput(input core.InputEvent) error {
	if input.Key == core.KeyBackspace {
		s.game.ChangeScene(s.game.playingScene)
	}
	return nil
}