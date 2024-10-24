package gameoflife

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/leaderboard"
	"github.com/kuhree/gg/internal/engine/render"
)

const (
	titleOffset = 1.0 / 10
	lineSpacing = 2
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
	cells              map[Vector2D]Cell
	playerPos          Vector2D
	prevLiveCellCount  int
	liveCellCount      int
	stableGenerations  int
	stableOscillations int
	boardStates        []uint64
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
	width, height := game.Size()
	playerPos := Vector2D{
		// X: float64(width) / 2,
		// Y: float64(height) / 2,
		X: 1,
		Y: 1,
	}

	scene := &PlayingScene{
		BaseScene: BaseScene{
			Game:          game,
			sceneName:     "Playing",
			blinkInterval: 0.5,
			showOnBlink:   true,
		},

		playerPos:          playerPos,
		cells:              make(map[Vector2D]Cell, height*width),
		prevLiveCellCount:  0,
		liveCellCount:      0,
		stableGenerations:  0,
		stableOscillations: 0,
		boardStates:        make([]uint64, 0, game.Config.StabilityThreshold),
	}

	return scene
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
		name:        "",
		nameEntered: false,
	}
}

// MainMenuScene methods

func (s *MainMenuScene) Draw(renderer *render.Renderer) {
	width, height := s.Size()
	startX := width / 10

	const (
		startOffset    = 1.0 / 6
		controlsOffset = 2.0 / 8
	)

	_ = renderer.DrawText(fmt.Sprintf("%s - %s", s.Config.Title, s.sceneName), startX, int(float64(height)*titleOffset), render.ColorWhite)

	if s.showOnBlink {
		_ = renderer.DrawText("Press ENTER to start", startX, int(float64(height)*startOffset), render.ColorBrightMagenta)
	}

	controlsY := int(float64(height) * controlsOffset)
	_ = renderer.DrawText("Controls:", startX, controlsY, render.ColorBlue)
	_ = renderer.DrawText("ESC to pause", startX, controlsY+3*lineSpacing, render.ColorWhite)
	_ = renderer.DrawText("Q to pause/quit", startX, controlsY+4*lineSpacing, render.ColorWhite)
}

func (s *MainMenuScene) HandleInput(input core.InputEvent) error {
	switch input.Rune {
	case core.KeyEnter:
		s.Logger.Info("Starting new game")
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

	s.updateCollisions(dt)
	s.checkGameState()
	s.CurrentLevel++
}

func (s *PlayingScene) Draw(renderer *render.Renderer) {
	width, height := s.Size()
	influencedCells := s.getPlayerInfluencedCells()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := s.getOrCreateCell(x, y)
			neighbors := s.countNeighbors(x, y)
			char, color := s.getCellInfo(float64(neighbors), float64(s.Config.BaseNeighboars))

			pos := Vector2D{X: float64(x), Y: float64(y)}
			isInfluenced := contains(influencedCells, pos)

			if c.Alive {
				if isInfluenced {
					_ = renderer.DrawRect(x, y, s.Config.BaseSize, s.Config.BaseSize, char, render.ColorBrightCyan)
				} else {
					_ = renderer.DrawRect(x, y, s.Config.BaseSize, s.Config.BaseSize, char, color)
				}
			} else {
				_ = renderer.DrawRect(x, y, s.Config.BaseSize, s.Config.BaseSize, ' ', render.ColorBrightBlack)
			}

			s.drawObjOverlay(x, y, c, color)
		}
	}

	// Draw player
	playerX := int(s.playerPos.X)
	playerY := int(s.playerPos.Y)
	_ = renderer.DrawRect(playerX, playerY, s.Config.BaseSize, s.Config.BaseSize, '@', render.ColorGreen)

	// Draw score, level, lives...
	_ = renderer.DrawText(fmt.Sprintf("Alive: %d", s.Score), 1, 1, render.ColorWhite)
	_ = renderer.DrawText(fmt.Sprintf("Level: %d", s.CurrentLevel), 1, 2, render.ColorWhite)
	if s.Overlay || s.Debug {
		_ = renderer.DrawText(fmt.Sprintf("Position: (%.1f, %.1f)", s.playerPos.X, s.playerPos.Y), 1, 3, render.ColorWhite)
		_ = renderer.DrawText(fmt.Sprintf("Stable Generations: %d / %d", s.stableGenerations, s.Config.StabilityThreshold), 1, 4, render.ColorWhite)
		_ = renderer.DrawText(fmt.Sprintf("Stable Oscillations: %d / %d", s.stableOscillations, s.Config.StabilityThreshold/2), 1, 5, render.ColorWhite)
	}
}

func (s *PlayingScene) HandleInput(input core.InputEvent) error {
	moveSpeed := s.Config.BaseMoveSpeed

	switch input.Rune {
	case '1', core.KeyF1:
		s.Debug = !s.Debug
	case '2', core.KeyF2:
		s.Overlay = !s.Overlay
	case core.KeyW:
		s.playerPos.Y -= moveSpeed
	case core.KeyS:
		s.playerPos.Y += moveSpeed
	case core.KeyA:
		s.playerPos.X -= moveSpeed
	case core.KeyD:
		s.playerPos.X += moveSpeed
	case 'p', 'P':
		s.Scenes.ChangeScene(PauseMenuSceneID)
	case 'q', 'Q':
		s.Scenes.ChangeScene(GameOverSceneID)
	}

	return nil
}

// PlayingScene helpers

func (s *PlayingScene) getOrCreateCell(x, y int) *Cell {
	pos := Vector2D{X: float64(x), Y: float64(y)}
	if cell, exists := s.cells[pos]; exists {
		return &cell
	}

	newCell := Cell{
		GameObject: GameObject{
			Position: pos,
			Width:    float64(s.Config.BaseSize),
			Height:   float64(s.Config.BaseSize),
		},
		Alive: rand.Float64() < s.Config.BaseChance,
	}
	s.cells[pos] = newCell
	return &newCell
}

// updateCollisions detects and handles collisions between game objects
func (s *PlayingScene) updateCollisions(_ float64) {
	width, height := s.Size()

	newCells := make(map[Vector2D]Cell)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := Vector2D{X: float64(x), Y: float64(y)}
			neighbors := s.countNeighbors(x, y)
			cell := s.getOrCreateCell(x, y)

			if cell.Alive && (neighbors < 2 || neighbors > 3) {
				newCells[pos] = Cell{Alive: false, GameObject: GameObject{Position: pos}}
			} else if !cell.Alive && neighbors == 3 {
				newCells[pos] = Cell{Alive: true, GameObject: GameObject{Position: pos}}
			} else {
				newCells[pos] = *cell
			}
		}
	}

	s.cells = newCells

	influencedCells := s.getPlayerInfluencedCells()
	radius := s.Config.BaseRadius
	for _, pos := range influencedCells {
		cell := s.getOrCreateCell(int(pos.X), int(pos.Y))
		distX := math.Abs(pos.X - s.playerPos.X)
		distY := math.Abs(pos.Y - s.playerPos.Y)
		if distX <= radius && distY <= radius {
			cell.Alive = true
		} else {
			// In the buffer zone, randomly activate cells
			cell.Alive = rand.Float64() < s.Config.BaseChance
		}
		s.cells[pos] = *cell
	}
}

// countNeighbors counts the number of live neighbors for a cell
func (s *PlayingScene) countNeighbors(x, y int) int {
	count := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i == 0 && j == 0 {
				continue
			}

			neighborX, neighborY := x+i, y+j
			pos := Vector2D{X: float64(neighborX), Y: float64(neighborY)}
			if cell, exists := s.cells[pos]; exists && cell.Alive {
				count++
			}
		}
	}

	return count
}

// checkGameState determines if the game should end
func (s *PlayingScene) checkGameState() {
	// Calculate and store the current board state hash
	currentHash := s.calculateBoardHash()
	s.boardStates = append(s.boardStates, currentHash)
	liveCells := 0
	width, height := s.Size()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := Vector2D{X: float64(x), Y: float64(y)}
			if cell, exists := s.cells[pos]; exists && cell.Alive {
				liveCells++
			}
		}
	}

	s.prevLiveCellCount = s.liveCellCount

	s.Score = liveCells
	s.liveCellCount = liveCells

	if s.liveCellCount == s.prevLiveCellCount {
		s.stableGenerations++
	} else {
		s.stableGenerations = 0
	}

	// Check for stable pattern
	ratio := float64(s.stableGenerations) / float64(s.Config.StabilityThreshold)
	if ratio >= s.Config.StabilityChance {
		s.endGame("Stability reached")
		return
	}

	// Check for oscillating patterns
	if len(s.boardStates) > s.Config.StabilityThreshold {
		s.boardStates = s.boardStates[1:] // Remove oldest state
	}

	s.stableOscillations = 0
	for _, hash := range s.boardStates[:len(s.boardStates)-1] {
		if hash == currentHash {
			s.stableOscillations++
		}
	}

	ratio = float64(s.stableOscillations) / float64(s.Config.StabilityThreshold/2)
	if ratio >= s.Config.StabilityChance {
		s.endGame("Oscillating pattern detected")
		return
	}

	// Increment generation count
	s.CurrentLevel++
}

func (s *PlayingScene) endGame(reason string) {
	s.Logger.Info("Game over", "reason", reason, "score", s.Score, "level", s.CurrentLevel+1)
	s.Scenes.ChangeScene(GameOverSceneID)
}

// calculateBoardHash computes a hash of the current board state
func (s *PlayingScene) calculateBoardHash() uint64 {
	var hash uint64
	influencedCells := s.getPlayerInfluencedCells()
	for pos, cell := range s.cells {
		if cell.Alive && !contains(influencedCells, pos) {
			hash ^= uint64(pos.X) * uint64(pos.Y)
		}
	}
	return hash
}

func contains(slice []Vector2D, item Vector2D) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func (s *PlayingScene) getCellInfo(neighbors float64, maxNeighbors float64) (rune, render.Color) {

	ratio := neighbors / float64(maxNeighbors)
	switch {
	case ratio >= 0.875:
		return render.FullBlock, render.ColorBrightRed
	case ratio >= 0.75:
		return render.DarkShade, render.ColorRed
	case ratio >= 0.625:
		return render.MediumShade, render.ColorYellow
	case ratio >= 0.5:
		return render.LightShade, render.ColorGreen
	case ratio >= 0.375:
		return '+', render.ColorCyan
	case ratio >= 0.25:
		return '*', render.ColorBlue
	case ratio >= 0.125:
		return '.', render.ColorMagenta
	default:
		return render.LightShade, render.ColorWhite
	}
}

// getPlayerInfluencedCells returns a list of cell positions influenced by the player
func (s *PlayingScene) getPlayerInfluencedCells() []Vector2D {
	influencedCells := make([]Vector2D, 0)
	width, height := s.Size()
	radius := s.Config.BaseRadius
	buffer := s.Config.BaseRadius * s.Config.SeedBuffer
	playerX, playerY := int(s.playerPos.X), int(s.playerPos.Y)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			distX := math.Abs(float64(x) - float64(playerX))
			distY := math.Abs(float64(y) - float64(playerY))
			if distX <= radius+buffer && distY <= radius+buffer {
				influencedCells = append(influencedCells, Vector2D{X: float64(x), Y: float64(y)})
			}
		}
	}

	return influencedCells
}

func (s *PlayingScene) drawObjOverlay(x, y int, cell *Cell, color render.Color) {
	if !s.Overlay && !s.Debug {
		return
	}

	if s.Overlay {
		char := '0'
		if cell.Alive {
			char = '1'
		}
		_ = s.Renderer.DrawChar(char, x, y, color)
	}

	if s.Debug {
		debugInfo := []string{
			fmt.Sprintf("P:%.1fX,%.1fY", cell.Position.X, cell.Position.Y),
			fmt.Sprintf("S:%.1fW,%.1fH", cell.Width, cell.Height),
		}
		for i, info := range debugInfo {
			_ = s.Renderer.DrawText(info, x, y+i, color)
		}
	}
}

// PauseMenuScene methods

func (s *PauseMenuScene) Draw(renderer *render.Renderer) {
	const (
		scoreOffset    = 1.0 / 6
		controlsOffset = 1.0 / 4
	)

	width, height := s.Size()
	startX := width / 10

	// Draw title
	_ = renderer.DrawText(fmt.Sprintf("%s - %s", s.Config.Title, s.sceneName), startX, int(float64(height)*titleOffset), render.ColorWhite)

	if s.showOnBlink {
		_ = renderer.DrawText(
			fmt.Sprintf("Score: %d | Level: %d", s.Score, s.CurrentLevel),
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
}

func (s *GameOverScene) GetDetails() string {
	width, height := s.Size()
	return fmt.Sprintf(
		"%dW*%dH|L%d@%.2fBC|%dBS|%.2fBR|%.2fSB|%dST",
		width, height,
		s.CurrentLevel, s.Config.BaseChance, s.Config.BaseSize,
		s.Config.BaseRadius, s.Config.SeedBuffer, s.Config.StabilityThreshold,
	)
}

func (s *GameOverScene) Draw(renderer *render.Renderer) {
	const (
		scoreOffset       = 1.0 / 6
		leaderboardOffset = 1.0 / 4
		controlsOffset    = 3.0 / 4
	)

	width, height := s.Size()
	startX := width / 10

	// Draw title and game over message
	_ = renderer.DrawText(fmt.Sprintf("%s - %s", s.Config.Title, s.sceneName), startX, int(float64(height)*titleOffset), render.ColorWhite)

	if s.Score > 0 && !s.nameEntered {
		// Draw name entry prompt and score
		_ = renderer.DrawText(fmt.Sprintf("Score: %d", s.Score), startX, int(float64(height)*scoreOffset), render.ColorWhite)
		_ = renderer.DrawText("Enter your name to save score (or press Q to skip):", startX, int(float64(height)*scoreOffset)+1, render.ColorWhite)
		if s.showOnBlink {
			_ = renderer.DrawText(s.name+"_", startX, int(float64(height)*scoreOffset)+2, render.ColorBrightMagenta)
		} else {
			_ = renderer.DrawText(s.name, startX, int(float64(height)*scoreOffset)+2, render.ColorBrightMagenta)
		}
	} else if s.Score > 0 && s.showOnBlink {
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
	_ = renderer.DrawText("Press ENTER to return to save/return to main menu", startX, controlsY+2*lineSpacing, render.ColorWhite)
}

func (s *GameOverScene) HandleInput(input core.InputEvent) error {
	switch input.Rune {
	case 'q', 'Q':
		if !s.nameEntered && s.Score > 0 {
			s.Logger.Info("Skipping leaderboard entry")
			s.nameEntered = true
		}
		return core.ErrQuitGame
	case core.KeyEnter:
		if !s.nameEntered {
			if len(s.name) > 0 && s.Score > 0 {
				s.nameEntered = true
				s.Logger.Info("Adding leaderboard entry...", "name", s.name, "score", s.Score)
				s.Leaderboard.Add(s.name, s.Score, s.GetDetails())
				err := s.Leaderboard.Save(s.Config.BoardFile)
				if err != nil {
					return err
				}
			}
		} else {
			s.Scenes.ChangeScene(MainMenuSceneID)
		}
	case core.KeyBackspace:
		if !s.nameEntered && len(s.name) > 0 {
			s.name = s.name[:len(s.name)-1]
		}
	default:
		if !s.nameEntered {
			// Only allow printable characters
			if input.Rune >= 32 && input.Rune <= 126 && len(s.name) < 12 {
				s.name += string(input.Rune)
			}
		}
	}
	return nil
}
