package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/kuhree/gg/examples/frames"
	"github.com/kuhree/gg/examples/space_invaders"
	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/utils"
)

var (
	height   int
	width    int
	debug    bool
	workDir  string
	gameName string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable Debug logging default")
	flag.StringVar(&gameName, "game", "", "Name of the game to launch")
	flag.StringVar(&workDir, "workDir", "./tmp", "Working directory for the game state")
	flag.IntVar(&width, "width", 80, "width of the game")
	flag.IntVar(&height, "height", 24, "height of the game")
}

func notImplemented() error {
	return fmt.Errorf("not yet implemented")
}

var games = []struct {
	name   string
	launch func() error
}{
	{"Frames", func() error {
		game := frames.NewGame(width, height)

		gameLoop := core.NewGameLoop(game)
		return gameLoop.Run()
	}},

	{"Space Invaders", func() error {
		game := space_invaders.NewGame(width, height, workDir, debug)

		gameLoop := core.NewGameLoop(game)
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
	utils.Logger.Info("Launching game", "name", gameName)

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
	showGameMenu()
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
