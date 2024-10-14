package space_invaders

import (
	"log/slog"
	"math"
	"time"

	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/render"
	"github.com/kuhree/gg/internal/engine/scenes"
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
	AliensCount   int     `json:"aliensCount"`
	AlienSpeed    float64 `json:"alienSpeed"`
	BarrierCount  int     `json:"barrierCount"`
	BarrierHealth int     `json:"barrierHealth"`
}

// Game represents the Space Invaders game state and logic
type Game struct {
	renderer     *render.Renderer
	logger       *slog.Logger
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


// StartNewGame initializes a new game
func (g *Game) StartNewGame() {
	g.logger.Info("Starting new game")
	g.score = 0
	g.currentLevel = 0
	g.player.lives = 3
	g.bulletSpeed = BulletSpeed
	g.initializeLevel()
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

	// Choose a formation based on the current level
	formation := g.currentLevel % 4

	alienWidth := float64(AlienSize)
	alienHeight := float64(AlienSize)

	var alienPositions []Vector2D

	switch formation {
	case 0:
		alienPositions = g.getRectangleFormation(levelData.AliensCount, alienWidth, alienHeight)
	case 1:
		alienPositions = g.getTriangleFormation(levelData.AliensCount, alienWidth, alienHeight)
	case 2:
		alienPositions = g.getDiamondFormation(levelData.AliensCount, alienWidth, alienHeight)
	case 3:
		alienPositions = g.getVFormation(levelData.AliensCount, alienWidth, alienHeight)
	}

	// Spawn aliens based on the calculated positions
	for i, pos := range alienPositions {
		if i >= levelData.AliensCount {
			break
		}
		var alienType AlienType
		if levelData.AliensCount >= 3 {
			alienType = AlienType(i / (levelData.AliensCount / 3))
		} else {
			alienType = AlienType(i % 3)
		}
		alien := &Alien{
			GameObject: GameObject{
				Position: pos,
				Speed:    levelData.AlienSpeed,
			},
			Type:   alienType,
			Points: (3 - int(alienType)) * 10,
		}
		g.aliens = append(g.aliens, alien)
	}

	// Setup barriers
	for i := 0; i < levelData.BarrierCount; i++ {
		barrier := &Barrier{
			GameObject: GameObject{
				Position: Vector2D{
					X: float64(i+1) * (float64(g.width) / (float64(levelData.BarrierCount) + 1)),
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
		"bulletSpeed", g.bulletSpeed,
		"formation", formation)
}

func (g *Game) getRectangleFormation(count int, alienWidth, alienHeight float64) []Vector2D {
	var positions []Vector2D
	rows := 5
	cols := count / rows
	if cols == 0 {
		cols = 1
	}

	startX := (float64(g.width) - float64(cols)*alienWidth) / 2
	startY := 2.0

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			positions = append(positions, Vector2D{
				X: startX + float64(col)*(alienWidth+1),
				Y: startY + float64(row)*(alienHeight+1),
			})
		}
	}

	return positions
}

func (g *Game) getTriangleFormation(count int, alienWidth, alienHeight float64) []Vector2D {
	var positions []Vector2D
	maxRows := 5
	maxCols := maxRows

	startX := (float64(g.width) - float64(maxCols)*alienWidth) / 2
	startY := 2.0

	for row := 0; row < maxRows && len(positions) < count; row++ {
		cols := row + 1
		for col := 0; col < cols && len(positions) < count; col++ {
			positions = append(positions, Vector2D{
				X: startX + float64(maxCols-cols)/2*alienWidth + float64(col)*(alienWidth+1),
				Y: startY + float64(row)*(alienHeight+1),
			})
		}
	}

	return positions
}

func (g *Game) getDiamondFormation(count int, alienWidth, alienHeight float64) []Vector2D {
	var positions []Vector2D
	maxRows := 7
	maxCols := 4

	startX := (float64(g.width) - float64(maxCols)*alienWidth) / 2
	startY := 2.0

	for row := 0; row < maxRows && len(positions) < count; row++ {
		cols := maxCols - abs(row-3)
		for col := 0; col < cols && len(positions) < count; col++ {
			positions = append(positions, Vector2D{
				X: startX + float64(maxCols-cols)/2*alienWidth + float64(col)*(alienWidth+1),
				Y: startY + float64(row)*(alienHeight+1),
			})
		}
	}

	return positions
}

func (g *Game) getVFormation(count int, alienWidth, alienHeight float64) []Vector2D {
	var positions []Vector2D
	maxRows := 5
	maxCols := maxRows*2 - 1

	startX := (float64(g.width) - float64(maxCols)*alienWidth) / 2
	startY := 2.0

	for row := 0; row < maxRows && len(positions) < count; row++ {
		cols := 2*row + 1
		for col := 0; col < cols && len(positions) < count; col++ {
			positions = append(positions, Vector2D{
				X: startX + float64(row)*alienWidth + float64(col)*(alienWidth+1),
				Y: startY + float64(row)*(alienHeight+1),
			})
		}
	}

	return positions
}

// Helper functions

func checkCollision(pos1, pos2 Vector2D, size1, size2 float64) bool {
	return int(math.Abs(float64(pos1.X-pos2.X))) < int((size1+size2)/2) && int(math.Abs(float64(pos1.Y-pos2.Y))) < int((size1+size2)/2)
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}