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

	alienPositions := g.getGroupedFormations(levelData.AliensCount, alienWidth, alienHeight)

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

// getGroupedFormations creates groups of mini formations and sprinkles in overflow
func (g *Game) getGroupedFormations(count int, alienWidth, alienHeight float64) []Vector2D {
	const (
		miniFormationSize = 4 // Size of each mini formation
		spacing           = 2.0 // Spacing between mini formations
	)

	var positions []Vector2D
	remainingAliens := count

	// Calculate number of complete mini formations
	numCompleteMiniFormations := remainingAliens / miniFormationSize

	// Calculate available width and height for formations
	formationWidth := 4*alienWidth + spacing
	formationHeight := 4*alienHeight + spacing

	// Calculate max formations per row and total rows
	formationsPerRow := int(float64(g.width) / formationWidth)
	if formationsPerRow == 0 {
		formationsPerRow = 1
	}
	totalRows := (numCompleteMiniFormations + formationsPerRow - 1) / formationsPerRow

	// Calculate start position to center the formations
	startX := (float64(g.width) - float64(formationsPerRow)*formationWidth + spacing) / 2
	startY := (float64(g.height) - float64(totalRows)*formationHeight) / 2

	for i := 0; i < numCompleteMiniFormations; i++ {
		formation := i % 4
		miniPositions := g.getMiniFormation(formation, miniFormationSize, alienWidth, alienHeight)

		// Calculate position for this mini formation
		col := i % formationsPerRow
		row := i / formationsPerRow
		offsetX := startX + float64(col) * formationWidth
		offsetY := startY + float64(row) * formationHeight

		// Adjust positions and add to main list
		for _, pos := range miniPositions {
			pos.X += offsetX
			pos.Y += offsetY
			positions = append(positions, pos)
		}

		remainingAliens -= miniFormationSize
	}

	// Handle overflow
	if remainingAliens > 0 {
		overflowStartX := startX + float64(numCompleteMiniFormations%formationsPerRow) * formationWidth
		overflowStartY := startY + float64(numCompleteMiniFormations/formationsPerRow) * formationHeight

		for i := 0; i < remainingAliens; i++ {
			col := i % 4
			row := i / 4
			positions = append(positions, Vector2D{
				X: overflowStartX + float64(col)*(alienWidth+1),
				Y: overflowStartY + float64(row)*(alienHeight+1),
			})
		}
	}

	return positions
}

// getMiniFormation returns a small formation of the specified type
func (g *Game) getMiniFormation(formationType, count int, alienWidth, alienHeight float64) []Vector2D {
	switch formationType {
	case 0:
		return g.getRectangleFormation(count, alienWidth, alienHeight)
	case 1:
		return g.getTriangleFormation(count, alienWidth, alienHeight)
	case 2:
		return g.getDiamondFormation(count, alienWidth, alienHeight)
	case 3:
		return g.getVFormation(count, alienWidth, alienHeight)
	default:
		return g.getRectangleFormation(count, alienWidth, alienHeight)
	}
}

// Modify existing formation functions to work with smaller counts
func (g *Game) getRectangleFormation(count int, alienWidth, alienHeight float64) []Vector2D {
	var positions []Vector2D
	rows := int(math.Min(float64(count/2+1), 2))
	cols := int(math.Min(float64(count/rows+1), 2))

	startX := -float64(cols-1) * (alienWidth + 1) / 2
	startY := -float64(rows-1) * (alienHeight + 1) / 2

	for row := 0; row < rows && len(positions) < count; row++ {
		for col := 0; col < cols && len(positions) < count; col++ {
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
	maxRows := 3

	for row := 0; row < maxRows && len(positions) < count; row++ {
		cols := row + 1
		startX := -float64(cols-1) * (alienWidth + 1) / 2
		startY := -float64(maxRows-1) * (alienHeight + 1) / 2

		for col := 0; col < cols && len(positions) < count; col++ {
			positions = append(positions, Vector2D{
				X: startX + float64(col)*(alienWidth+1),
				Y: startY + float64(row)*(alienHeight+1),
			})
		}
	}

	return positions
}

func (g *Game) getDiamondFormation(count int, alienWidth, alienHeight float64) []Vector2D {
	var positions []Vector2D
	maxRows := 3
	maxCols := 2

	for row := 0; row < maxRows && len(positions) < count; row++ {
		cols := maxCols - abs(row-1)
		startX := -float64(cols-1) * (alienWidth + 1) / 2
		startY := -float64(maxRows-1) * (alienHeight + 1) / 2

		for col := 0; col < cols && len(positions) < count; col++ {
			positions = append(positions, Vector2D{
				X: startX + float64(col)*(alienWidth+1),
				Y: startY + float64(row)*(alienHeight+1),
			})
		}
	}

	return positions
}

func (g *Game) getVFormation(count int, alienWidth, alienHeight float64) []Vector2D {
	var positions []Vector2D
	maxRows := 2
	maxCols := 3

	for row := 0; row < maxRows && len(positions) < count; row++ {
		cols := 2*row + 1
		startX := -float64(maxCols-1) * (alienWidth + 1) / 2
		startY := -float64(maxRows-1) * (alienHeight + 1) / 2

		for col := 0; col < cols && len(positions) < count; col++ {
			positions = append(positions, Vector2D{
				X: startX + float64(row+col)*(alienWidth+1),
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