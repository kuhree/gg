package gameoflife

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

// Game represents the Space Invaders game state and logic

type Game struct {
	// Internal engine stuff
	Height      int
	Width       int
	Renderer    *render.Renderer
	Logger      *slog.Logger
	Scenes      *scenes.Manager
	Leaderboard *leaderboard.Board

	// Game-specific ui/state/debugging
	Config  *Config
	Debug   bool
	Overlay bool

	// Game-specific state
	Score        int
	CurrentLevel int
}

// NewGame creates a new instance of the game
func NewGame(width, height int, workDir string, debug bool, overlay bool) (*Game, error) {
	logger := utils.Logger
	renderer := render.NewRenderer(width, height, render.DefaultPalette)
	scenes := scenes.NewManager()

	config, err := NewConfig(workDir)
	if err != nil {
		return nil, err
	} else {
		logger.Info("Config loaded!", "path", config.ConfigFile, "config", config)
	}

	logger.Debug(config.BoardFile)
	board, err := leaderboard.NewBoard(config.BoardFile)
	if err != nil {
		return nil, err
	} else {
		logger.Info("Board loaded!", "path", config.BoardFile, "board", board)
	}

	game := &Game{
		Height:      height,
		Width:       width,
		Renderer:    renderer,
		Logger:      logger,
		Config:      config,
		Leaderboard: board,
		Debug:       debug,
		Overlay:     overlay,
		Scenes:      scenes,
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
	if g.Score > 0 {
		err := g.Leaderboard.Save(g.Config.BoardFile)
		if err != nil {
			g.Logger.Error(fmt.Sprintf("%s - Leaderboard failed to save", g.Config.Title), "err", err)
		} else {
			g.Logger.Info("Leaderboard entry saved!...", "score", g.Score)
		}
	}

	g.Logger.Info(fmt.Sprintf("%s - Game cleaned up", g.Config.Title))
}

func (g *Game) Size() (int, int) {
	return g.Width, g.Height
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
