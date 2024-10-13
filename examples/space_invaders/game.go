package space_invaders

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/kuhree/gg/examples/space_invaders/scenes"
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
	sceneManager *scenes.Manager

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
		bulletSpeed: BulletSpeed,
	}

	game.sceneManager = scenes.NewManager()
	game.sceneManager.AddScene(scenes.MainMenuSceneID, NewMainMenuScene(game))
	game.sceneManager.AddScene(scenes.PlayingSceneID, NewPlayingScene(game))
	game.sceneManager.AddScene(scenes.GameOverSceneID, NewGameOverScene(game))
	game.sceneManager.AddScene(scenes.PauseMenuSceneID, NewPauseMenuScene(game))

	game.sceneManager.ChangeScene(scenes.MainMenuSceneID)

	if err := game.LoadLevels("examples/space_invaders/levels.json"); err != nil {
		logger.Error("Failed to load levels", "error", err)
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
	g.logger.Info("Space Invaders game cleaned up")
}

// Draw renders the game state
func (g *Game) Draw() {
	g.renderer.Clear()
	g.sceneManager.Draw(g.renderer)
	g.renderer.Render()
}

// Update updates the game state
func (g *Game) Update(dt float64) error {
	g.sceneManager.Update(dt)
	return nil
}

// HandleInput processes user input
func (g *Game) HandleInput(input core.InputEvent) error {
	return g.sceneManager.HandleInput(input)
}


// FireBullet creates a new bullet from the player's position
func (g *Game) FireBullet() {
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

// StartNewGame initializes a new game
func (g *Game) StartNewGame() {
	g.logger.Info("Starting new game")
	g.score = 0
	g.currentLevel = 0
	g.player.lives = 3
	g.bulletSpeed = BulletSpeed
	g.initializeLevel()
}

// MovePlayer updates the player's position based on the given direction
func (g *Game) MovePlayer(dx, dy int) {
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
