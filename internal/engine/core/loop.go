package core

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kuhree/gg/internal/engine/render"
	"github.com/kuhree/gg/internal/utils"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/term"
)

// GameLoop manages the main game loop
type GameLoop struct {
	game         Game
	term         *term.Terminal
	termOldState *term.State
	logger       *slog.Logger
	running      bool
	keyEvents    chan InputEvent

	resize  chan os.Signal
	signals chan os.Signal
}

// NewGameLoop creates a new GameLoop
func NewGameLoop(game Game) *GameLoop {
	gl := &GameLoop{
		game:      game,
		logger:    utils.Logger,
		keyEvents: make(chan InputEvent, 1), // Buffer for key events
		resize:    make(chan os.Signal, 1),
		signals:   make(chan os.Signal, 1),
	}

	return gl
}

// Run starts the game loop
func (gl *GameLoop) Run(targetTime, targetFps float64) error {
	gl.running = true
	err := gl.game.Init()
	if err != nil {
		return err
	}

	// Set terminal into raw mode to capture input
	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer terminal.Restore(fd, oldState)
	defer render.ShowCursor()

	gl.term = term.NewTerminal(os.Stdin, "")
	gl.updateTerminalSize(gl.term)
	gl.term.SetBracketedPasteMode(true)

	// Capture signals to gracefully exit
	signal.Notify(gl.signals, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(gl.resize, syscall.SIGWINCH)

	// Start listener in a separate goroutine
	go func() {
		for gl.running {
			var buf [1]byte

			n, err := os.Stdin.Read(buf[:])
			if err != nil {
				if err.Error() == "EOF" {
					gl.logger.Error("EOF received, exiting input loop.")
					close(gl.keyEvents) // Close the channel on EOF
					return
				}

				gl.logger.Error("Error reading from stdin", "err", err)
				close(gl.keyEvents) // Close the channel on error
				return
			}

			if n > 0 {
				gl.logger.Debug("Key pressed", "key", fmt.Sprintf("%c", buf), "rune", rune(buf[0]))
				gl.keyEvents <- InputEvent{
					Rune: rune(buf[0]),
				}
			}
		}
	}()

	targetDeltaTime := 1.0 / targetFps
	lastTime := time.Now()
	for gl.running {
		currentTime := time.Now()
		deltaTime := currentTime.Sub(lastTime).Seconds()
		deltaTime *= targetTime // Speedup/slowdown the game
		lastTime = currentTime

		// Handle input (non-blocking)
		select {
		case <-gl.resize:
			gl.updateTerminalSize(gl.term)
		case <-gl.signals:
			gl.logger.Info("Signal Received. Exiting...")
			gl.Stop()
		case keyEvent, ok := <-gl.keyEvents:
			if !ok {
				gl.logger.Error("Unable to access keyboard channel. Exiting", "err", err)
				gl.Stop()
				continue
			} else if err := gl.game.HandleInput(keyEvent); err != nil {
				gl.Stop()
				if err != ErrQuitGame {
					gl.logger.Error("Game failed to update. Exiting..", "err", err)
					return err
				}
			}
		default:
			// No input, continue with the game loop
		}

		// Update game state
		err := gl.game.Update(deltaTime)
		if err != nil {
			gl.Stop()
			if err != ErrQuitGame {
				gl.logger.Error("Game failed to update. Exiting..", "err", err)
				return err
			}
		}

		// Render
		gl.game.Draw()

		// Cap the frame rate
		sleepTime := targetDeltaTime - deltaTime
		if sleepTime > 0 {
			time.Sleep(time.Duration(sleepTime * float64(time.Second)))
		}
	}

	gl.game.Cleanup()
	return nil
}

// Stop stops the game loop
func (gl *GameLoop) Stop() {
	gl.running = false
}

func (gl *GameLoop) updateTerminalSize(term *term.Terminal) {
}
