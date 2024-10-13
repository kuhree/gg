package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/kuhree/gg/examples/frames"
	"github.com/kuhree/gg/examples/space_invaders"
	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/render"
	"github.com/kuhree/gg/internal/utils"
)

var (
	debug    bool
	gameName string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable Debug logging default")
	flag.StringVar(&gameName, "game", "", "Name of the game to launch")
}

func notImplemented() error {
	return fmt.Errorf("not yet implemented")
}

var games = []struct {
	name   string
	launch func() error
}{
	{"Frames", func() error {
		renderer := render.NewRenderer(80, 24) // Create a 80x24 ASCII renderer
		game := frames.NewGame(renderer, utils.Logger)

		gameLoop := core.NewGameLoop(game, renderer, utils.Logger)
		return gameLoop.Run()
	}},

	{"Space Invaders", func() error {
		renderer := render.NewRenderer(80, 24) // Create a 80x24 ASCII renderer
		game := space_invaders.NewGame(renderer, utils.Logger, )

		gameLoop := core.NewGameLoop(game, renderer, utils.Logger)
		return gameLoop.Run()
	}},
	{"Pong", notImplemented},
	{"Tetris", notImplemented},
	{"Pac-Man", notImplemented},
	{"Snake", notImplemented},
}

func main() {
	flag.Parse()
	if gameName == "" {
		gameName = flag.Arg(0)
	}

	if debug {
		utils.SetLogLevel(slog.LevelDebug)
	}

	utils.Logger.Info("Starting GG (Go Game Engine)", "debug", debug)

	if gameName != "" {
		launchGame(gameName)
	} else {
		showGameMenu()
	}
}

func launchGame(gameName string) {
	utils.Logger.Info("Launching game", "name", gameName, "devMode", debug)

	for _, game := range games {
		if game.name == gameName {
			err := game.launch()
			if err != nil {
				utils.Logger.Error("Failed to launch game", "name", gameName, "error", err)
				return
			}
			return
		}
	}

	utils.Logger.Error("Game not found", "name", gameName)
}

func showGameMenu() {
	utils.Logger.Info("Showing game selection menu")

	for i, game := range games {
		fmt.Printf("%d. %s\n", i+1, game.name)
	}

	var choice int
	for {
		fmt.Print("Enter the number of the game you want to play (or 0 to quit): ")
		_, err := fmt.Scanf("%d", &choice)
		if err != nil || choice < 0 || choice > len(games) {
			fmt.Println("Invalid input. Please try again.")
			continue
		}
		break
	}

	if choice == 0 {
		utils.Logger.Info("Exiting game selection")
		os.Exit(0)
	}

	selectedGame := games[choice-1]
	utils.Logger.Info("Game selected", "name", selectedGame.name)
	err := selectedGame.launch()
	if err != nil {
		utils.Logger.Error("Failed to launch game", "error", err)
	}
}
