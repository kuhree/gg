package core

import (
	"log/slog"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/kuhree/gg/internal/engine/render"
)

// GameLoop manages the main game loop
type GameLoop struct {
	game     Game
	renderer *render.Renderer
	logger   *slog.Logger
	running  bool
	keyEvents chan keyboard.KeyEvent
}

// NewGameLoop creates a new GameLoop
func NewGameLoop(game Game, renderer *render.Renderer, logger *slog.Logger) *GameLoop {
	return &GameLoop{
		game:     game,
		renderer: renderer,
		logger:   logger,
		keyEvents: make(chan keyboard.KeyEvent, 10), // Buffer for key events
	}
}

// Run starts the game loop
func (gl *GameLoop) Run() error {
	gl.running = true
	err := gl.game.Init()
	if err != nil {
		return err
	}

	// Start keyboard listener in a separate goroutine
	go gl.listenForKeyboard()

	const targetFPS = 60
	const targetDeltaTime = 1.0 / targetFPS

	lastTime := time.Now()

	for gl.running {
		currentTime := time.Now()
		deltaTime := currentTime.Sub(lastTime).Seconds()
		lastTime = currentTime

		// Handle input (non-blocking)
		select {
		case keyEvent := <-gl.keyEvents:
			err := gl.game.HandleInput(InputEvent{KeyEvent: keyEvent})
			if err == ErrQuitGame {
				gl.running = false
			} else if err != nil {
				return err
			}
		default:
			// No input, continue with the game loop
		}

		// Update game state
		err := gl.game.Update(deltaTime)
		if err != nil {
			if err == ErrQuitGame {
				gl.running = false
			} else {
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

// listenForKeyboard continuously listens for keyboard input
func (gl *GameLoop) listenForKeyboard() {
	err := keyboard.Open()
	if err != nil {
		gl.logger.Error("Failed to open keyboard", "error", err)
		return
	}
	defer keyboard.Close()

	for gl.running {
		char, key, err := keyboard.GetKey()
		if err != nil {
			gl.logger.Error("Error getting key", "error", err)
			continue
		}
		gl.keyEvents <- keyboard.KeyEvent{
			Key:  key,
			Rune: char,
			Err:  err,
		}
	}
}

// Stop stops the game loop
func (gl *GameLoop) Stop() {
	gl.running = false
}
