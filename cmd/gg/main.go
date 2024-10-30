// Implements a retro-style game launcher for the GG game engine.
// It provides a command-line interface to launch various ASCII-based games.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kuhree/gg/examples/breakout"
	"github.com/kuhree/gg/examples/flappybird"
	"github.com/kuhree/gg/examples/frames"
	"github.com/kuhree/gg/examples/gameoflife"
	"github.com/kuhree/gg/examples/spaceinvaders"
	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/utils"
)

// Game settings
var (
	// Display settings
	width  int
	height int
	fps    float64

	// Game engine settings
	time    float64
	workDir string

	// Debug settings
	debug   bool
	overlay bool
)

// CLI flags
var (
	gameName  string
	listGames bool
)

// Game represents a playable game in the collection
type Game struct {
	Name        string
	Description string
	Launch      func() error
}

func init() {
	flag.BoolVar(&listGames, "list", false, "List all available games")
	flag.StringVar(&gameName, "game", "", "Name/Index of the game to launch")
	flag.StringVar(&workDir, "workDir", getDefaultWorkDir(), "Working directory for the game state")
	flag.IntVar(&width, "width", 80, "width of the game")
	flag.IntVar(&height, "height", 24, "height of the game")
	flag.Float64Var(&time, "time", 1.0, "target time elapse within game")
	flag.Float64Var(&fps, "fps", 60, "target fps within game (24,30,60,120,240)")

	flag.BoolVar(&overlay, "overlay", false, "Enable some useful overlays")
	flag.BoolVar(&debug, "debug", false, "Enable Debug logging. Will enable all other debug attributes.")
}

func main() {
	flag.Parse()

	if listGames {
		fmt.Println("Available games:")
		for i, game := range games {
			fmt.Printf("%d. %s: %s\n", i+1, game.Name, game.Description)
		}
		os.Exit(0)
	}

	if gameName == "" {
		gameName = flag.Arg(0)
	}

	if debug {
		overlay = true
		if err := utils.SetLogLevel(slog.LevelDebug); err != nil {
			utils.Logger.Error("Failed to set log level", "error", err)
			os.Exit(1)
		}

		if err := utils.SetLogFile(filepath.Join(getDefaultWorkDir(), "gg.log")); err != nil {
			utils.Logger.Error("Failed to set log file", "error", err)
			os.Exit(1)
		}
	}

	utils.Logger.Info("Starting GG", "debug", debug)
	defer func() {
		_ = utils.Cleanup()
	}()

	if gameName != "" {
		launchGame(gameName)
	} else {
		showGameMenu()
	}
}

// Follow XDG Base Directory Specification
// https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
func getDefaultWorkDir() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "./tmp"
		}
		dataHome = filepath.Join(homeDir, ".local", "share")
	}

	dataDir := filepath.Join(dataHome, "gg")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "./tmp"
	}
	return dataDir
}

func launchSelectedGame(game Game) {
	utils.Logger.Info("Game selected", "name", game.Name)
	err := game.Launch()
	if err != nil {
		utils.Logger.Error("Failed to launch game", "error", err)
	}
}

func launchGame(gameName string) {
	utils.Logger.Info("Launching game", "name", gameName)

	for i, game := range games {
		if game.Name == gameName {
			launchSelectedGame(game)
			return
		}

		// Fallback to id/index
		gameId, err := strconv.Atoi(gameName)
		if err == nil && i+1 == gameId {
			launchSelectedGame(game)
			return
		}
	}

	utils.Logger.Error("Game not found", "name", gameName)
	showGameMenu()
}

func showGameMenu() {
	utils.Logger.Info("Showing game selection menu")

	for i, game := range games {
		fmt.Printf("%d. %s: %s\n", i+1, game.Name, game.Description)
	}

	var choice string
	for {
		fmt.Print("Enter the number or name of the game you want to play (or 'q' to quit): ")
		_, err := fmt.Scanln(&choice)
		if err != nil {
			utils.Logger.Error("Failed to get game selection", "err", err)
			os.Exit(1)
		}

		if choice == "q" {
			utils.Logger.Info("Exiting game selection")
			os.Exit(0)
		}

		// Try to parse as number
		if num, err := strconv.Atoi(choice); err == nil && num > 0 && num <= len(games) {
			launchSelectedGame(games[num-1])
			return
		}

		// Try to match by name
		for _, game := range games {
			if strings.EqualFold(choice, game.Name) {
				launchSelectedGame(game)
				return
			}
		}

		fmt.Println("Invalid input. Please try again.")
	}
}

var games = []Game{
	{
		"Frames",
		"A technical demo showcasing the game engine's core rendering capabilities and performance",
		func() error {
			game := frames.NewGame(width, height)

			gl := core.NewGameLoop(game)
			if err := gl.Run(time, fps); err != nil {
				return err
			}
			defer gl.Stop()

			return nil
		},
	},

	{
		"Space Invaders",
		"Classic arcade shooter - defend Earth from waves of descending aliens in this timeless game",
		func() error {
			game, err := space_invaders.NewGame(width, height, workDir, debug, overlay)
			if err != nil {
				return err
			}

			gl := core.NewGameLoop(game)
			if err := gl.Run(time, fps); err != nil {
				return err
			}
			defer gl.Stop()

			return nil
		},
	},

	{
		"Game of Life",
		"Conway's famous cellular automaton simulation - watch patterns emerge from simple rules",
		func() error {
			game, err := gameoflife.NewGame(width, height, workDir, debug, overlay)
			if err != nil {
				return err
			}

			gl := core.NewGameLoop(game)
			if err := gl.Run(time, fps); err != nil {
				return err
			}
			defer gl.Stop()

			return nil
		},
	},
	{
		"Breakout",
		"Arcade classic - break blocks and chase high scores in this addictive paddle game",
		func() error {
			game, err := breakout.NewGame(width, height, workDir, debug, overlay)
			if err != nil {
				return err
			}

			gl := core.NewGameLoop(game)
			if err := gl.Run(time, fps); err != nil {
				return err
			}
			defer gl.Stop()

			return nil
		},
	},
	{
		"Flappy Bird",
		"Modern classic - navigate through pipes with precise timing in this challenging side-scroller",
		func() error {
			game, err := flappybird.NewGame(width, height, workDir, debug, overlay)
			if err != nil {
				return err
			}

			gl := core.NewGameLoop(game)
			if err := gl.Run(time, fps); err != nil {
				return err
			}
			defer gl.Stop()

			return nil
		},
	},
}
