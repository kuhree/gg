package core

// GameState represents the current state of the game
type GameState int

const (
	StateMainMenu GameState = iota
	StateGameplay
	StatePaused
	StateGameOver
)

// StateManager handles transitions between different game states
type StateManager struct {
	currentState GameState
}

// NewStateManager creates a new StateManager
func NewStateManager() *StateManager {
	return &StateManager{
		currentState: StateMainMenu,
	}
}

// GetCurrentState returns the current game state
func (sm *StateManager) GetCurrentState() GameState {
	return sm.currentState
}

// SetState changes the current game state
func (sm *StateManager) SetState(state GameState) {
	sm.currentState = state
}
