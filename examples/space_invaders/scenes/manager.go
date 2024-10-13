package scenes

import (
	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/render"
)

type SceneID int

const (
	MainMenuSceneID SceneID = iota
	PlayingSceneID
	GameOverSceneID
	PauseMenuSceneID
)

// Add this interface definition
type Scene interface {
	Enter()
	Exit()
	Update(dt float64)
	Draw(renderer *render.Renderer)
	HandleInput(input core.InputEvent) error
}

type Manager struct {
	scenes       map[SceneID]Scene
	currentScene Scene
}

// Update the NewManager function
func NewManager() *Manager {
	return &Manager{
		scenes: make(map[SceneID]Scene),
	}
}

func (m *Manager) AddScene(id SceneID, scene Scene) {
	m.scenes[id] = scene
}

func (m *Manager) ChangeScene(id SceneID) {
	if m.currentScene != nil {
		m.currentScene.Exit()
	}
	m.currentScene = m.scenes[id]
	m.currentScene.Enter()
}

func (m *Manager) Update(dt float64) {
	if m.currentScene != nil {
		m.currentScene.Update(dt)
	}
}

func (m *Manager) Draw(renderer *render.Renderer) {
	if m.currentScene != nil {
		m.currentScene.Draw(renderer)
	}
}

func (m *Manager) HandleInput(input core.InputEvent) error {
	if m.currentScene != nil {
		return m.currentScene.HandleInput(input)
	}
	return nil
}
