package core

// Key constants
const (
	KeyF1 = rune(999) // Escape sequence for F1
	KeyF2 = rune(998) // Escape sequence for F2
	// KeyF3        = rune('\x1b') // Escape sequence for F3
	// KeyF4        = rune('\x1b') // Escape sequence for F4
	// KeyDelete    = rune('\x1b') // Escape sequence for Delete
	// KeyUp        = rune('\x1b') // Escape sequence for Up arrow
	// KeyDown      = rune('\x1b') // Escape sequence for Down arrow
	// KeyLeft      = rune('\x1b') // Escape sequence for Left arrow
	// KeyRight     = rune('\x1b') // Escape sequence for Right arrow
	KeyTab       = rune('\t')
	KeyEnter     = rune('\r')
	KeyEscape    = rune('\x1b')
	KeySpace     = rune(' ')
	KeyBackspace = rune(127)

	KeyQ = rune(113)
	KeyE = rune(101)
	KeyW = rune(119)
	KeyA = rune(97)
	KeyS = rune(115)
	KeyD = rune(100)
)

// InputEvent represents an input event from the user
type InputEvent struct {
	Rune rune
}
