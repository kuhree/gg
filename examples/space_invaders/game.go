package space_invaders

import (
	"fmt"
	"log/slog"

	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/leaderboard"
	"github.com/kuhree/gg/internal/engine/render"
	"github.com/kuhree/gg/internal/engine/scenes"
	"github.com/kuhree/gg/internal/utils"
)

const (
	MainMenuSceneID scenes.SceneID = iota
	PlayingSceneID
	PauseMenuSceneID
	GameOverSceneID
)

var BaseConfig = Config{
	Title:      "Space Invaders",
	GameDir:    "spaceinvaders",
	ConfigFile: "config.json",
	BoardFile:  "board.json",

	PlayerYOffset:  3,
	BarrierYOffset: 7,
	AlienYOffset:   3,

	BaseScore:                0,
	BaseLives:                3,
	BaseLevel:                1,
	BaseLevelStep:            1,
	BaseDifficulty:           1.0,
	BaseDifficultyMultiplier: 0.1,

	BasePlayerSize:   2.0,
	BasePlayerSpeed:  1.0,
	BasePlayerHealth: 10.0,
	BasePlayerAttack: 5.0,

	BaseAliensCount: 1,
	BaseAlienSize:   2.0,
	BaseAlienSpeed:  1.0,
	BaseAlienHealth: 20.0,

	BaseProjectileSize:   2.0,
	BaseProjectileSpeed:  30.0,
	BaseProjectileHealth: 10.0,

	BaseBarrierCount:  10,
	BaseBarrierSize:   2.0,
	BaseBarrierHealth: 100.0,
	BaseBarrierAttack: 0.0,

	BaseShootInterval:     15.0,
	MinShootInterval:      5.0,
	ShootIntervalVariance: 20.0,
	BaseShootChance:       0.2,
	CooldownMultiplier:    1.5,
	IntervalRandomFactor:  0.5,
}

// Game represents the Space Invaders game state and logic
type Game struct {
	// Internal engine stuff
	Renderer    *render.Renderer
	Logger      *slog.Logger
	Scenes      *scenes.Manager
	Leaderboard *leaderboard.Board

	// Game-specific ui/state/debugging
	Config *Config
	Debug  bool

	// Game-specific state
	Score        int
	CurrentLevel int

	// Game-specific objects
	Player            *Player
	Aliens            []*Alien
	Projectiles       []*Projectile
	Barriers          []*Barrier
	BarriersCountLast int
}

// NewGame creates a new instance of the Space Invaders game
func NewGame(width, height int, workDir string, debug bool) (*Game, error) {
	logger := utils.Logger
	renderer := render.NewRenderer(width, height)
	scenes := scenes.NewManager()

	config, err := NewConfig(workDir, &BaseConfig)
	if err != nil {
		return nil, err
	} else {
		logger.Info("Config loaded!", "path", config.ConfigFile, "config", config)
	}

	board, err := leaderboard.NewBoard(config.BoardFile)
	if err != nil {
		return nil, err
	} else {
		logger.Info("Board loaded!", "path", config.BoardFile, "board", board)
	}

	game := &Game{
		Renderer:    renderer,
		Logger:      logger,
		Config:      config,
		Leaderboard: board,
		Debug:       debug,
		Scenes:      scenes,
		Player: &Player{
			GameObject: GameObject{
				Position: Vector2D{X: float64(width) / 2, Y: float64(height) - float64(config.PlayerYOffset)},
				Speed:    Vector2D{X: config.BasePlayerSpeed, Y: config.BasePlayerSpeed},
				Health:   config.BasePlayerHealth,
				Attack:   config.BasePlayerAttack,
				Width:    config.BasePlayerSize,
				Height:   config.BasePlayerSize,
			},

			Lives: config.BaseLives,
		},
	}

	return game, nil
}

// Init initializes the game
func (g *Game) Init() error {
	g.Logger.Info(fmt.Sprintf("%s - Game initializing...", g.Config.Title))

	g.Logger.Info(fmt.Sprintf("%s - Adding Scenes", g.Config.Title))
	g.Scenes.AddScene(MainMenuSceneID, NewMainMenuScene(g))
	g.Scenes.AddScene(PlayingSceneID, NewPlayingScene(g))
	g.Scenes.AddScene(GameOverSceneID, NewGameOverScene(g))
	g.Scenes.AddScene(PauseMenuSceneID, NewPauseMenuScene(g))
	g.Scenes.ChangeScene(MainMenuSceneID)
	g.Logger.Info(fmt.Sprintf("%s - Scenes loaded!", g.Config.Title), "startScene", MainMenuSceneID)

	g.Logger.Info(fmt.Sprintf("%s - Game initialized", g.Config.Title))
	return nil
}

// Cleanup performs any necessary cleanup
func (g *Game) Cleanup() {
	err := g.Leaderboard.Save(g.Config.BoardFile)
	if err != nil {
		g.Logger.Error(fmt.Sprintf("%s - Leaderboard failed to save", g.Config.Title), "err", err)
		return
	}

	g.Logger.Info(fmt.Sprintf("%s - Game cleaned up", g.Config.Title))
}

// Draw renders the game state
func (g *Game) Draw() {
	g.Renderer.Clear()
	g.Scenes.Draw(g.Renderer)
	g.Renderer.Render()
}

// Update updates the game state
func (g *Game) Update(dt float64) error {
	g.Scenes.Update(dt)
	return nil
}

// HandleInput processes user input
func (g *Game) HandleInput(input core.InputEvent) error {
	return g.Scenes.HandleInput(input)
}
