package space_invaders

import (
	"fmt"
	"math"

	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/leaderboard"
	"github.com/kuhree/gg/internal/engine/render"
)

// BaseScene provides common functionality for all scenes
type BaseScene struct {
	*Game
	sceneName     string
	blinkTimer    float64
	blinkInterval float64
	showOnBlink   bool
}

// Enter logs when a scene is entered
func (s *BaseScene) Enter() {
	s.Logger.Info("Entering scene", "scene", s.sceneName)
}

// Exit logs when a scene is exited
func (s *BaseScene) Exit() {
	s.Logger.Info("Exiting scene", "scene", s.sceneName)
}

// Update is a no-op for scenes that don't need updates
func (s *BaseScene) Update(dt float64) {
	s.blinkTimer += dt
	if s.blinkTimer >= s.blinkInterval {
		s.blinkTimer = 0
		s.showOnBlink = !s.showOnBlink
	}
}

// HandleInput is a no-op for scenes that don't handle input
func (s *BaseScene) HandleInput(input core.InputEvent) error {
	return nil
}

// MainMenuScene represents the main menu
type MainMenuScene struct {
	BaseScene
}

// PlayingScene represents the main gameplay
type PlayingScene struct {
	BaseScene
	collectableSpawnTimer float64
}

// PauseMenuScene represents the pause menu
type PauseMenuScene struct {
	BaseScene
}

// GameOverScene represents the game over screen
type GameOverScene struct {
	BaseScene
	name        string
	nameEntered bool
}

// NewMainMenuScene creates a new main menu scene
func NewMainMenuScene(game *Game) *MainMenuScene {
	return &MainMenuScene{
		BaseScene: BaseScene{
			Game:          game,
			sceneName:     "Main Menu",
			blinkInterval: 0.5,
			showOnBlink:   true,
		},
	}
}

// NewPlayingScene creates a new playing scene
func NewPlayingScene(game *Game) *PlayingScene {
	return &PlayingScene{
		BaseScene: BaseScene{
			Game:          game,
			sceneName:     "Playing",
			blinkInterval: 0.5,
			showOnBlink:   true,
		},
	}
}

// NewPauseMenuScene creates a new pause menu scene
func NewPauseMenuScene(game *Game) *PauseMenuScene {
	return &PauseMenuScene{
		BaseScene: BaseScene{
			Game:          game,
			sceneName:     "Pause Menu",
			blinkInterval: 0.5,
			showOnBlink:   true,
		},
	}
}

// NewGameOverScene creates a new game over scene
func NewGameOverScene(game *Game) *GameOverScene {
	return &GameOverScene{
		BaseScene: BaseScene{
			Game:          game,
			sceneName:     "Game Over",
			blinkInterval: 0.5,
			showOnBlink:   true,
		},
	}
}

// MainMenuScene methods

func (s *MainMenuScene) Draw(renderer *render.Renderer) {
	width, height := s.Size()
	startX := width / 10

	const (
		titleOffset    = 1.0 / 10
		startOffset    = 1.0 / 6
		controlsOffset = 2.0 / 8
		lineSpacing    = 2
	)

	_ = renderer.DrawText(fmt.Sprintf("%s - %s", s.Config.Title, s.sceneName), startX, int(float64(height)*titleOffset), render.ColorWhite)

	if s.showOnBlink {
		_ = renderer.DrawText("Press ENTER to start", startX, int(float64(height)*startOffset), render.ColorBrightMagenta)
	}

	controlsY := int(float64(height) * controlsOffset)
	_ = renderer.DrawText("Controls:", startX, controlsY, render.ColorBlue)
	_ = renderer.DrawText("Arrow keys / WASD to move", startX, controlsY+lineSpacing, render.ColorWhite)
	_ = renderer.DrawText("SPACE to shoot", startX, controlsY+2*lineSpacing, render.ColorWhite)
	_ = renderer.DrawText("ESC to pause", startX, controlsY+3*lineSpacing, render.ColorWhite)
	_ = renderer.DrawText("Q to pause/quit", startX, controlsY+4*lineSpacing, render.ColorWhite)
}

func (s *MainMenuScene) HandleInput(input core.InputEvent) error {
	switch input.Rune {
	case core.KeyEnter:
		s.Logger.Info("Starting new game")
		s.CurrentLevel = s.Config.BaseLevel - s.Config.BaseLevelStep
		s.Score = s.Config.BaseScore
		s.Player.Health = s.Config.BasePlayerHealth
		s.Player.MaxHealth = s.Config.BasePlayerHealth
		s.Player.Lives = s.Config.BaseLives
		s.Scenes.ChangeScene(PlayingSceneID)
		return nil
	case 'q', 'Q':
		s.Scenes.ChangeScene(GameOverSceneID)
		return core.ErrQuitGame
	}

	return nil
}

// PlayingScene methods

func (s *PlayingScene) Update(dt float64) {
	s.BaseScene.Update(dt)
	s.updateCollectables(dt)
	s.updateAliens(dt)
	s.updateProjectiles(dt)
	s.updateCollisions()

	s.murder()
	s.updateGameState()
}

func (s *PlayingScene) Draw(renderer *render.Renderer) {
	// Draw player
	player := s.Player
	playerChar, playerColor := s.getHealthInfo(player.Health, s.Player.MaxHealth)

	_ = s.Renderer.DrawRect(
		int(player.Position.X-player.Width/2),
		int(player.Position.Y-player.Height/2),
		int(player.Width),
		int(player.Height),
		playerChar,
		playerColor,
	)

	s.drawObjOverlay(&player.GameObject, playerColor, OverlayOpts{})

	// Draw collectibles
	for _, collectable := range s.Collectables {
		char, color := s.getCollectableInfo(collectable)

		_ = s.Renderer.DrawRect(
			int(collectable.Position.X-collectable.Width/2),
			int(collectable.Position.Y-collectable.Height/2),
			int(collectable.Width),
			int(collectable.Height),
			char,
			color,
		)
	}

	// Draw aliens
	for _, alien := range s.Aliens {
		char, color := s.getAlienInfo(alien)

		if s.Debug {
			_ = s.Renderer.DrawText(fmt.Sprintf("%d", alien.AlienType),
				int(alien.Position.X-alien.Width/2),
				int(alien.Position.Y-alien.Height/2)-1,
				color,
			)
		}

		_ = s.Renderer.DrawRect(
			int(alien.Position.X-alien.Width/2),
			int(alien.Position.Y-alien.Height/2),
			int(alien.Width),
			int(alien.Height),
			char,
			color,
		)

		s.drawObjOverlay(&alien.GameObject, color, OverlayOpts{})
	}

	// Draw projectiles
	for _, projectile := range s.Projectiles {
		char, color := s.getProjectileInfo(projectile)
		_ = s.Renderer.DrawRect(
			int(projectile.Position.X-projectile.Width/2),
			int(projectile.Position.Y-projectile.Height/2),
			int(projectile.Width),
			int(projectile.Height),
			char,
			color,
		)

		s.drawObjOverlay(&projectile.GameObject, render.ColorWhite, OverlayOpts{})
	}

	// Draw barriers
	for _, barrier := range s.Barriers {
		char, color := s.getBarrierInfo(barrier.Health, barrier.MaxHealth)

		_ = s.Renderer.DrawRect(
			int(barrier.Position.X-barrier.Width/2),
			int(barrier.Position.Y-barrier.Height/2),
			int(barrier.Width),
			int(barrier.Height),
			char,
			color,
		)

		s.drawObjOverlay(&barrier.GameObject, render.ColorWhite, OverlayOpts{Health: true})
	}

	// Draw score, level, lives...
	info := []struct {
		format string
		args   []interface{}
		color  render.Color
	}{
		{"Score: %d", []interface{}{s.Score}, render.ColorWhite},
		{"Level: %d", []interface{}{s.CurrentLevel}, render.ColorWhite},
		{"Enemies: %d", []interface{}{len(s.Aliens)}, render.ColorWhite},
		{"Health: %.2f", []interface{}{player.Health}, playerColor},
		{"Lives: %d", []interface{}{player.Lives}, render.ColorWhite},
	}

	for i, item := range info {
		_ = s.Renderer.DrawText(fmt.Sprintf(item.format, item.args...), 1, i+1, item.color)
	}

}

type OverlayOpts struct {
	Health bool
}

func (s *PlayingScene) drawObjOverlay(obj *GameObject, color render.Color, opts OverlayOpts) {
	_, healthColor := s.getHealthInfo(obj.Health, obj.MaxHealth)

	if s.Overlay || opts.Health {
		_ = s.Renderer.DrawText(
			fmt.Sprintf("%.f", math.Round(obj.Health)),
			int(obj.Position.X-obj.Width/2),
			int(obj.Position.Y+obj.Height/2)-1,
			healthColor,
		)
	}

	if s.Debug {
		debugInfo := []struct {
			format string
			args   []interface{}
		}{
			{"P:%.fX,%.fY", []interface{}{obj.Position.X, obj.Position.Y}},
			{"A:%.fWx%.fH", []interface{}{obj.Width, obj.Height}},
			{"S:%.fX,%.fY", []interface{}{obj.Speed.X, obj.Speed.Y}},
		}

		for i, info := range debugInfo {
			_ = s.Renderer.DrawText(
				fmt.Sprintf(info.format, info.args...),
				int(obj.Position.X+obj.Width/2)+1,
				int(obj.Position.Y-obj.Height/2)-1+i,
				color,
			)
		}
	}
}

func (s *PlayingScene) HandleInput(input core.InputEvent) error {
	switch input.Rune {
	case '1', core.KeyF1:
		s.Debug = !s.Debug
	case '2', core.KeyF2:
		s.Overlay = !s.Overlay
	case core.KeyEscape, core.KeyTab, 'q', 'Q', 'p', 'P':
		s.Scenes.ChangeScene(PauseMenuSceneID)
	case 'w', 'W':
		s.movePlayer(0, -1)
	case 'a', 'A':
		s.movePlayer(-1, 0)
	case 's', 'S':
		s.movePlayer(0, 1)
	case 'd', 'D':
		s.movePlayer(2, 0)
	case core.KeySpace:
		s.shoot(&s.Player.GameObject)
	case '_':
		s.CurrentLevel -= s.Config.BaseLevelStep
		s.startWave()
	case '+':
		s.CurrentLevel += s.Config.BaseLevelStep
		s.startWave()
	}

	return nil
}

// PauseMenuScene methods

func (s *PauseMenuScene) Draw(renderer *render.Renderer) {
	const (
		titleOffset    = 1.0 / 10
		scoreOffset    = 1.0 / 6
		controlsOffset = 1.0 / 4
		lineSpacing    = 2
	)

	width, height := s.Size()
	startX := width / 10

	// Draw title
	_ = renderer.DrawText(fmt.Sprintf("%s - %s", s.Config.Title, s.sceneName), startX, int(float64(height)*titleOffset), render.ColorWhite)

	if s.showOnBlink {
		_ = renderer.DrawText(
			fmt.Sprintf("Score: %d | Level: %d | Enemies Remaining | %d", s.Score, s.CurrentLevel, len(s.Aliens)),
			startX,
			int(float64(height)*scoreOffset),
			render.ColorMagenta,
		)
	}

	// Draw controls
	controlsY := int(float64(height) * controlsOffset)
	_ = renderer.DrawText("Controls:", startX, controlsY, render.ColorBlue)
	_ = renderer.DrawText("Press ESC to resume", startX, controlsY+lineSpacing, render.ColorWhite)
	_ = renderer.DrawText("Press Q to quit", startX, controlsY+2*lineSpacing, render.ColorWhite)
}

func (s *PauseMenuScene) HandleInput(input core.InputEvent) error {
	switch input.Rune {
	case core.KeyEscape:
		s.Scenes.ChangeScene(PlayingSceneID)
	case 'q', 'Q':
		s.Scenes.ChangeScene(GameOverSceneID)
		return core.ErrQuitGame
	}

	return nil
}

// GameOverScene methods

func (s *GameOverScene) Enter() {
	s.BaseScene.Enter()
	err := s.Leaderboard.Load(s.Config.BoardFile)
	if err != nil {
		s.Logger.Warn("Failed to load existing leaderboard. Creating a new one...", "path", s.Config.BoardFile, "err", err)
		s.Leaderboard.Records = make([]leaderboard.Record, 0)
	}

	s.Logger.Info("Adding leaderboard entry...", "score", s.Score)
	s.Leaderboard.Add(
		"anon",
		s.Score,
		s.GetDetails(),
	)
}

func (s *GameOverScene) GetDetails() string {
	width, height := s.Size()
	return fmt.Sprintf(
		"%dW*%dH|L%d@%dBL|%.1fBH|%.1fBAH|(%.2fBD * %.1fBDM)|%dBS",
		width, height,
		s.CurrentLevel, s.Config.BaseLevel,
		s.Config.BasePlayerHealth, s.Config.BaseAlienHealth,
		s.Config.BaseDifficulty, s.Config.BaseDifficultyMultiplier, s.Config.BaseScore,
	)
}

func (s *GameOverScene) Draw(renderer *render.Renderer) {
	const (
		titleOffset       = 1.0 / 10
		scoreOffset       = 1.0 / 6
		leaderboardOffset = 1.0 / 4
		controlsOffset    = 3.0 / 4
		lineSpacing       = 2
	)

	width, height := s.Size()
	startX := width / 10

	// Draw title and game over message
	_ = renderer.DrawText(fmt.Sprintf("%s - %s", s.Config.Title, s.sceneName), startX, int(float64(height)*titleOffset), render.ColorWhite)

	if !s.nameEntered {
		// Draw name entry prompt
		_ = renderer.DrawText("Enter your name:", startX, int(float64(height)*scoreOffset), render.ColorWhite)
		if s.showOnBlink {
			_ = renderer.DrawText(s.name+"_", startX, int(float64(height)*scoreOffset)+2, render.ColorBrightMagenta)
		} else {
			_ = renderer.DrawText(s.name, startX, int(float64(height)*scoreOffset)+2, render.ColorBrightMagenta)
		}
		return
	}

	if s.showOnBlink {
		_ = renderer.DrawText(
			fmt.Sprintf("%d | %s > %s", s.Score, s.name, s.GetDetails()),
			startX,
			int(float64(height)*scoreOffset),
			render.ColorMagenta,
		)
	}

	// Draw leaderboard
	leaderboardY := int(float64(height) * leaderboardOffset)
	_ = renderer.DrawText("Top Scores:", startX, leaderboardY, render.ColorBlue)
	topScores := s.Leaderboard.TopScores(5)
	for i, entry := range topScores {
		_ = renderer.DrawText(fmt.Sprintf("%d | %s > %s", entry.Score, entry.Name, entry.Details), startX, leaderboardY+(i+1)*lineSpacing, render.ColorWhite)
	}

	// Draw controls
	controlsY := int(float64(height) * controlsOffset)
	_ = renderer.DrawText("Controls:", startX, controlsY, render.ColorBlue)
	_ = renderer.DrawText("Press Q to quit the game", startX, controlsY+lineSpacing, render.ColorWhite)
	_ = renderer.DrawText("Press ENTER to return to main menu", startX, controlsY+2*lineSpacing, render.ColorWhite)
}

func (s *GameOverScene) HandleInput(input core.InputEvent) error {
	if input.Rune == core.KeyEnter {
		s.Scenes.ChangeScene(MainMenuSceneID)
		return nil
	}

	switch input.Rune {
	case 'q', 'Q':
		return core.ErrQuitGame
	}
	return nil
}
