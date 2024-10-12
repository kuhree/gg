package core

import (
	"github.com/eiannone/keyboard"
	"github.com/kuhree/gg/internal/utils"
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
	KeyLeft  = keyboard.KeyArrowLeft
	KeyRight = keyboard.KeyArrowRight
	KeySpace = keyboard.KeySpace
	KeyEnter = keyboard.KeyEnter
	KeyBackspace     = keyboard.KeyBackspace
)

// InputEvent represents an input event from the user
type InputEvent struct {
	Type InputEventType
	Key  keyboard.Key
	X    int
	Y    int
}

// InputHandler manages input events
type InputHandler struct {
	events []InputEvent
}

// NewInputHandler creates a new InputHandler
func NewInputHandler() *InputHandler {
	return &InputHandler{
		events: make([]InputEvent, 0),
	}
}

// AddEvent adds a new input event to the handler
func (ih *InputHandler) AddEvent(event InputEvent) {
	ih.events = append(ih.events, event)
}

// PollEvents returns all current events and clears the event queue
func (ih *InputHandler) PollEvents() []InputEvent {
	events := ih.events
	ih.events = make([]InputEvent, 0)
	return events
}

// Handle input (this is a placeholder, you'll need to implement actual input handling)
// For example, you might use a library like "github.com/eiannone/keyboard"
// to read keyboard input non-blockingly
func (ih *InputHandler) Scan(keyEvents <-chan keyboard.KeyEvent, quit func()) error {
	select {
	case event := <-keyEvents:
		if event.Err != nil {
			return event.Err
		}

		utils.Logger.Debug("Key Pressed", "Key", event.Key, "Rune", event.Rune)

		if event.Key == keyboard.KeyEsc {
			quit()
			return nil
		}

		ih.AddEvent(InputEvent{
			Type: KeyPress,
			Key:  event.Key,
		})
	default:
		// No input, continue with the game loop
	}

	return nil
}
