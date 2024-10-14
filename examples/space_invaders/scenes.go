package space_invaders

import (
	"fmt"
	"math"

	"github.com/kuhree/gg/internal/engine/core"
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
}

// PauseMenuScene represents the pause menu
type PauseMenuScene struct {
	BaseScene
}

// GameOverScene represents the game over screen
type GameOverScene struct {
	BaseScene
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
	width, height := s.Renderer.Size()
	startX := width / 10

	const (
		titleOffset    = 1.0 / 10
		startOffset    = 1.0 / 6
		controlsOffset = 2.0 / 8
		lineSpacing    = 2
	)

	_ = renderer.DrawText(fmt.Sprintf("%s - %s", TITLE, s.sceneName), startX, int(float64(height)*titleOffset), render.ColorWhite)

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
	switch input.Key {
	case core.KeyEnter:
		s.Logger.Info("Starting new game")
		s.CurrentLevel = s.Config.BaseLevel
		s.Score = s.Config.BaseScore
		s.Player.Health = s.Config.BasePlayerHealth
		s.Player.MaxHealth = s.Config.BasePlayerHealth
		s.Player.Lives = s.Config.BaseLives
		s.Scenes.ChangeScene(PlayingSceneID)
		return nil
	default:
		switch input.Rune {
		case 'q', 'Q':
			s.Scenes.ChangeScene(GameOverSceneID)
			return core.ErrQuitGame
		}
	}

	return nil
}

// PlayingScene methods

func (s *PlayingScene) Enter() {
	s.BaseScene.Enter()
	s.startWave()
}

func (s *PlayingScene) Exit() {
	s.BaseScene.Exit()
}

func (s *PlayingScene) Update(dt float64) {
	s.BaseScene.Update(dt)
	s.updateAliens(dt)
	s.updateAlienFiringSquad(dt)
	s.updateProjectiles(dt)
	s.updateCollisions()

	aliensCountPrev := len(s.Aliens) // must be before removing/killing, ehhh
	s.killTheDead()
	s.updateGameState(aliensCountPrev)
}

func (s *PlayingScene) Draw(renderer *render.Renderer) {
	width, _ := s.Renderer.Size()

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

	s.drawObjOverlay(&player.GameObject, playerColor, OverlayOpts{Health: s.Debug, Attack: s.Debug})

	// Draw aliens
	for _, alien := range s.Aliens {
		alienChar, alienColor := s.getAlienInfo(alien)

		if s.Debug {
			_ = s.Renderer.DrawText(fmt.Sprintf("%d", alien.AlienType),
				int(alien.Position.X-alien.Width/2),
				int(alien.Position.Y-alien.Height/2)-1,
				alienColor,
			)
		}

		_ = s.Renderer.DrawRect(
			int(alien.Position.X-alien.Width/2),
			int(alien.Position.Y-alien.Height/2),
			int(alien.Width),
			int(alien.Height),
			alienChar,
			alienColor,
		)

		s.drawObjOverlay(&alien.GameObject, alienColor, OverlayOpts{Health: true, Attack: s.Debug})
	}

	// Draw projectiles
	for _, projectile := range s.Projectiles {
		projectileChar, projectileColor := s.getProjectileInfo(projectile)
		_ = s.Renderer.DrawRect(
			int(projectile.Position.X-projectile.Width/2),
			int(projectile.Position.Y-projectile.Height/2),
			int(projectile.Width),
			int(projectile.Height),
			projectileChar,
			projectileColor,
		)

		s.drawObjOverlay(&projectile.GameObject, render.ColorWhite, OverlayOpts{Health: s.Debug, Attack: true})
	}

	// Draw barriers
	for _, barrier := range s.Barriers {
		barrierChar, barrierColor := s.getBarrierInfo(barrier.Health, barrier.MaxHealth)

		_ = s.Renderer.DrawRect(
			int(barrier.Position.X-barrier.Width/2),
			int(barrier.Position.Y-barrier.Height/2),
			int(barrier.Width),
			int(barrier.Height),
			barrierChar,
			barrierColor,
		)

		s.drawObjOverlay(&barrier.GameObject, render.ColorWhite, OverlayOpts{Health: true, Attack: s.Debug})
	}

	// Draw score, level, lives...
	_ = s.Renderer.DrawText(fmt.Sprintf("Score: %d", s.Score), 1, 1, render.ColorWhite)
	_ = s.Renderer.DrawText(fmt.Sprintf("Level: %d", s.CurrentLevel), 1, 2, render.ColorWhite)
	_ = s.Renderer.DrawText(fmt.Sprintf("Enemies: %d", len(s.Aliens)), 1, 3, render.ColorWhite)

	_ = s.Renderer.DrawText(fmt.Sprintf("Health: %.2f", player.Health), width-13, 1, playerColor)
	_ = s.Renderer.DrawText(fmt.Sprintf("Attack: %.2f", player.Attack), width-12, 2, render.ColorWhite)
	_ = s.Renderer.DrawText(fmt.Sprintf("Lives: %d", player.Lives), width-8, 3, render.ColorWhite)
}

type OverlayOpts struct {
	Health bool
	Attack bool
}

func (s *PlayingScene) drawObjOverlay(obj *GameObject, color render.Color, opts OverlayOpts) {
	_, healthColor := s.getHealthInfo(obj.Health, obj.MaxHealth)

	if opts.Health {
		_ = s.Renderer.DrawText(
			fmt.Sprintf("%.f", math.Round(obj.Health)),
			int(obj.Position.X-obj.Width/2),
			int(obj.Position.Y+obj.Height/2)-1,
			healthColor,
		)
	}

	if opts.Attack {
		_ = s.Renderer.DrawText(
			fmt.Sprintf("%.f", math.Round(obj.Attack)),
			int(obj.Position.X-obj.Width/2),
			int(obj.Position.Y-obj.Height/2),
			render.ColorRed,
		)
	}

	if s.Debug {
		_ = s.Renderer.DrawText(
			fmt.Sprintf("P:%.fX,%.fY", obj.Position.X, obj.Position.Y),
			int(obj.Position.X+obj.Width/2)+1,
			int(obj.Position.Y-obj.Height/2)-1,
			color,
		)
		_ = s.Renderer.DrawText(
			fmt.Sprintf("A:%.fWx%.fH", obj.Width, obj.Height),
			int(obj.Position.X+obj.Width/2)+1,
			int(obj.Position.Y-obj.Height/2),
			color,
		)
		_ = s.Renderer.DrawText(
			fmt.Sprintf("S:%.fX,%.fY", obj.Speed.X, obj.Speed.Y),
			int(obj.Position.X+obj.Width/2)+1,
			int(obj.Position.Y-obj.Height/2)+1,
			color,
		)
	}
}

func (s *PlayingScene) HandleInput(input core.InputEvent) error {
	s.Logger.Info("Key pressed", "key", input.Key, "rune", input.Rune)

	switch input.Key {
	case core.KeySpace:
		s.Shoot(&s.Player.GameObject)
	case core.KeyLeft:
		s.movePlayer(-1, 0)
	case core.KeyRight:
		s.movePlayer(1, 0)
	case core.KeyUp:
		s.movePlayer(0, -1)
	case core.KeyDown:
		s.movePlayer(0, 1)
	case core.KeyEscape:
		s.Scenes.ChangeScene(PauseMenuSceneID)
	default:
		switch input.Rune {
		case 'q', 'Q':
			s.Scenes.ChangeScene(PauseMenuSceneID)
		case 'w', 'W':
			s.movePlayer(0, -1)
		case 'a', 'A':
			s.movePlayer(-1, 0)
		case 's', 'S':
			s.movePlayer(0, 1)
		case 'd', 'D':
			s.movePlayer(1, 0)
		case ' ':
			s.Shoot(&s.Player.GameObject)
		}
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

	width, height := s.Renderer.Size()
	startX := width / 10

	// Draw title
	_ = renderer.DrawText(fmt.Sprintf("%s - %s", TITLE, s.sceneName), startX, int(float64(height)*titleOffset), render.ColorWhite)

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
	_ = renderer.DrawText("Press R to Restart", startX, controlsY+3*lineSpacing, render.ColorWhite)
}

func (s *PauseMenuScene) HandleInput(input core.InputEvent) error {
	switch input.Key {
	case core.KeyEscape:
		s.Scenes.ChangeScene(PlayingSceneID)
	default:
		switch input.Rune {
		case 'q', 'Q':
			s.Scenes.ChangeScene(GameOverSceneID)
			return core.ErrQuitGame
		case 'r', 'R':
			s.Scenes.ChangeScene(PlayingSceneID)
		}
	}

	return nil
}

// GameOverScene methods

func (s *GameOverScene) Enter() {
	s.BaseScene.Enter()

	if s.Score > 0 {
		err := s.Leaderboard.Load(s.Config.BoardFile)
		if err != nil {
			s.Logger.Error("Failed to load leaderboard", "err", err)
			return
		}

		width, height := s.Renderer.Size()
		s.Logger.Info("Adding leaderboard entry...", "score", s.Score)
		s.Leaderboard.Add("anon", s.Score, fmt.Sprintf("Level: %d, Health: %.2f, Attack: %.2f | (%dx%d)", s.CurrentLevel, s.Player.Health, s.Player.Attack, width, height))
	}
}

func (s *GameOverScene) Draw(renderer *render.Renderer) {
	const (
		titleOffset       = 1.0 / 10
		scoreOffset       = 1.0 / 6
		leaderboardOffset = 1.0 / 4
		controlsOffset    = 3.0 / 4
		lineSpacing       = 2
	)

	width, height := s.Renderer.Size()
	startX := width / 10

	// Draw title and game over message
	_ = renderer.DrawText(fmt.Sprintf("%s - %s", TITLE, s.sceneName), startX, int(float64(height)*titleOffset), render.ColorWhite)
	if s.showOnBlink {
		_ = renderer.DrawText(
			fmt.Sprintf("Score: %d | Level: %d", s.Score, s.CurrentLevel),
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
		_ = renderer.DrawText(fmt.Sprintf("%d. %s: %d | %s", i+1, entry.Name, entry.Score, entry.Notes), startX, leaderboardY+(i+1)*lineSpacing, render.ColorWhite)
	}

	// Draw controls
	controlsY := int(float64(height) * controlsOffset)
	_ = renderer.DrawText("Controls:", startX, controlsY, render.ColorBlue)
	_ = renderer.DrawText("Press Q to quit the game", startX, controlsY+lineSpacing, render.ColorWhite)
	_ = renderer.DrawText("Press ENTER to return to main menu", startX, controlsY+2*lineSpacing, render.ColorWhite)
}

func (s *GameOverScene) HandleInput(input core.InputEvent) error {
	if input.Key == core.KeyEnter {
		s.Scenes.ChangeScene(MainMenuSceneID)
		return nil
	}

	switch input.Rune {
	case 'q', 'Q':
		return core.ErrQuitGame
	}
	return nil
}
