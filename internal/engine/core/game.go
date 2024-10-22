package core

import (
	"errors"
)

// Game interface defines the methods that all games must implement
type Game interface {
	Init() error
	Update(dt float64) error
	Size() (int, int)
	Draw()
	HandleInput(InputEvent) error
	Cleanup()
}

// Common errors
var (
	ErrQuitGame = errors.New("quit game")
)
