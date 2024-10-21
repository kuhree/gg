package core

import (
	"github.com/eiannone/keyboard"
)

// InputEventType represents the type of input event
type InputEventType int

const (
	KeyPress InputEventType = iota
	KeyRelease
	MousePress
	MouseRelease
	MouseMove
)

// Key constants
const (
	KeyUp        = keyboard.KeyArrowUp
	KeyDown      = keyboard.KeyArrowDown
	KeyLeft      = keyboard.KeyArrowLeft
	KeyRight     = keyboard.KeyArrowRight
	KeySpace     = keyboard.KeySpace
	KeyEnter     = keyboard.KeyEnter
	KeyBackspace = keyboard.KeyBackspace
	KeyEscape    = keyboard.KeyEsc
	KeyTab       = keyboard.KeyTab

	KeyQ = rune(113)
	KeyE = rune(101)
	KeyW = rune(119)
	KeyA = rune(97)
	KeyS = rune(115)
	KeyD = rune(100)
)

// InputEvent represents an input event from the user
type InputEvent struct {
	keyboard.KeyEvent
	X int
	Y int
}
