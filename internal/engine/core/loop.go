package core

import (
	"log/slog"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/kuhree/gg/internal/engine/render"
)

// GameLoop manages the main game loop
type GameLoop struct {
	game         Game
	renderer     *render.Renderer
	inputHandler *InputHandler
	logger       *slog.Logger
	running      bool
}

// NewGameLoop creates a new GameLoop
func NewGameLoop(game Game, renderer *render.Renderer, inputHandler *InputHandler, logger *slog.Logger) *GameLoop {
	return &GameLoop{
		game:     game,
		renderer: renderer,
		logger:   logger,
		inputHandler: inputHandler,
	}
}

// Run starts the game loop
func (gl *GameLoop) Run() error {
	gl.running = true
	err := gl.game.Init()
	if err != nil {
		return err
	}

	const targetFPS = 60
	const targetDeltaTime = 1.0 / targetFPS

	lastTime := time.Now()

	keyEvents, err := keyboard.GetKeys(10)
	if err != nil {
		return err
	}

	for gl.running {
		currentTime := time.Now()
		deltaTime := currentTime.Sub(lastTime).Seconds()
		lastTime = currentTime

		// Update game state
		err := gl.game.Update(deltaTime)
		if err != nil {
			if err == ErrQuitGame {
				gl.running = false
			} else {
				return err
			}
		}

		err = gl.inputHandler.Scan(keyEvents, func() { gl.running = false })
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

// Stop stops the game loop
func (gl *GameLoop) Stop() {
	gl.running = false
}
