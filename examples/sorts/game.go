package sorts

import (
	"fmt"
	"log/slog"

	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/render"
	"github.com/kuhree/gg/internal/engine/scenes"
	"github.com/kuhree/gg/internal/utils"
)

const (
	MainMenuSceneID scenes.SceneID = iota
	VisualizerSceneID
)

type Game struct {
	Height   int
	Width    int
	Renderer *render.Renderer
	Logger   *slog.Logger
	Scenes   *scenes.Manager
	Config   *Config
	Debug    bool
	Overlay  bool

	// Sorting specific state
	CurrentArray    []int
	CurrentSorter   Sorter
	SortComplete    bool
	ComparisonCount int
	SwapCount       int
	ElapsedTime     float64
}

func NewGame(width, height int, workDir string, debug bool, overlay bool) (*Game, error) {
	logger := utils.Logger
	renderer := render.NewRenderer(width, height, render.DefaultPalette)
	scenes := scenes.NewManager()

	config, err := NewConfig(workDir)
	if err != nil {
		return nil, err
	}

	game := &Game{
		Height:    height,
		Width:     width,
		Renderer:  renderer,
		Logger:    logger,
		Config:    config,
		Debug:     debug,
		Overlay:   overlay,
		Scenes:    scenes,
		SortComplete: false,
	}

	return game, nil
}

func (g *Game) Init() error {
	g.Logger.Info(fmt.Sprintf("%s - Game initializing...", g.Config.Title))

	g.Logger.Info(fmt.Sprintf("%s - Adding Scenes", g.Config.Title))
	g.Scenes.AddScene(MainMenuSceneID, NewMainMenuScene(g))
	g.Scenes.AddScene(VisualizerSceneID, NewVisualizerScene(g))
	g.Scenes.ChangeScene(MainMenuSceneID)

	g.Logger.Info(fmt.Sprintf("%s - Game initialized", g.Config.Title))
	return nil
}

func (g *Game) Cleanup() {
	g.Logger.Info(fmt.Sprintf("%s - Game cleaned up", g.Config.Title))
}

func (g *Game) Size() (int, int) {
	return g.Width, g.Height
}

func (g *Game) Draw() {
	g.Renderer.Clear()
	g.Scenes.Draw(g.Renderer)
	g.Renderer.Render()
}

func (g *Game) Update(dt float64) error {
	g.Scenes.Update(dt)
	return nil
}

func (g *Game) HandleInput(input core.InputEvent) error {
	return g.Scenes.HandleInput(input)
}
