package space_invaders

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/kuhree/gg/internal/engine/core"
	"github.com/kuhree/gg/internal/engine/render"
)

// GameState represents the current state of the game
type GameState int

const (
	MainMenu GameState = iota
	Playing
	GameOver
	PauseMenu
)

// AlienType represents different types of aliens
type AlienType int

const (
	LowAlien AlienType = iota
	MediumAlien
	HighAlien
)

// Vector2D represents a 2D position or velocity
type Vector2D struct {
	X, Y float64
}

// Entity represents a basic game entity with position and speed
type Entity struct {
	Position Vector2D
	Speed    float64
}

// Player represents the player's ship
type Player struct {
	Entity

	lives int
}

// Alien represents an enemy alien
type Alien struct {
	Entity
	Type   AlienType
	Points int
}

// Bullet represents a projectile fired by the player or aliens
type Bullet struct {
	Entity
	Damage int
}

// Barrier represents a defensive structure
type Barrier struct {
	Entity
	Health int
}

// Game represents the Space Invaders game state and logic
type Game struct {
	renderer     *render.Renderer
	logger       *slog.Logger
	inputHandler *core.InputHandler
	state        GameState

	score    int
	level    int
	lastTime time.Time

	player   *Player
	aliens   []*Alien
	bullets  []*Bullet
	barriers []*Barrier

	width  int
	height int

	// Main menu fields
	blinkTimer     float64
	blinkInterval  float64
	showPressEnter bool

	// Game over fields
	displayScore int

	// Pause menu fields
	pausePulseTimer float64
	pausePulseScale float64
}

const (
	PlayerSpeed  = 200
	BulletSpeed  = 300
	AlienSpeed   = 50
	PlayerSize   = 30
	AlienSize    = 30
	BulletSize   = 5
	BarrierSize  = 60
)

// NewGame creates a new instance of the Space Invaders game
func NewGame(renderer *render.Renderer, logger *slog.Logger, inputHandler *core.InputHandler) *Game {
	width, height := renderer.Size()

	return &Game{
		renderer:     renderer,
		logger:       logger,
		inputHandler: inputHandler,
		state:        MainMenu, // Start at MainMenu instead of Playing
		lastTime:     time.Now(),
		width:        width,
		height:       height,
		player: &Player{
			Entity: Entity{
				Position: Vector2D{X: float64(width) / 2, Y: float64(height) - 50},
				Speed:    PlayerSpeed,
			},
			lives: 3,
		},
		blinkInterval:  0.5, // Blink every 0.5 seconds
		showPressEnter: true,
	}
}

// Init initializes the game
func (g *Game) Init() error {
	g.logger.Info("Space Invaders game initialized")
	return nil
}

// Cleanup performs any necessary cleanup
func (g *Game) Cleanup() {
	g.logger.Info("Space Invaders game cleaned up")
}

// Draw renders the game state
func (g *Game) Draw() {
	g.logger.Debug("Drawing game state", "state", g.state)
	g.renderer.Clear()
	switch g.state {
	case MainMenu:
		g.drawMainMenu()
	case Playing:
		g.drawPlaying()
	case GameOver:
		g.drawGameOver()
	case PauseMenu:
		g.drawPauseMenu()
	}
	g.renderer.Render()
}

// drawMainMenu handles the drawing logic for the main menu
func (g *Game) drawMainMenu() {
	g.renderer.DrawText("SPACE INVADERS", g.width/2, g.height/3)
	if g.showPressEnter {
		g.renderer.DrawText("Press ENTER to start", g.width/2, g.height/2)
	}
}

// drawPlaying handles the drawing logic for the playing state
func (g *Game) drawPlaying() {
	// Draw player
	g.renderer.DrawRect(int(g.player.Position.X), int(g.player.Position.Y), int(PlayerSize), int(PlayerSize), '+')

	// Draw aliens
	for _, alien := range g.aliens {
		g.renderer.DrawRect(int(alien.Position.X), int(alien.Position.Y), int(AlienSize), int(AlienSize), '+')
	}

	// Draw bullets
	for _, bullet := range g.bullets {
		g.renderer.DrawRect(int(bullet.Position.X), int(bullet.Position.Y), int(BulletSize), int(BulletSize), '+')
	}

	// Draw barriers
	for _, barrier := range g.barriers {
		g.renderer.DrawRect(int(barrier.Position.X), int(barrier.Position.Y), int(BarrierSize), int(BarrierSize), '+')
	}

	// Draw score and lives
	g.renderer.DrawText(fmt.Sprintf("Score: %d", g.score), 10, 10)
	g.renderer.DrawText(fmt.Sprintf("Lives: %d", g.player.lives), g.width-10, 10)
}

// drawGameOver handles the drawing logic for the game over screen
func (g *Game) drawGameOver() {
	g.renderer.DrawText("GAME OVER", g.width/2, g.height/3)
	g.renderer.DrawText(fmt.Sprintf("Final Score: %d", g.displayScore), g.width/2, g.height/2)
	g.renderer.DrawText("Press ENTER to restart", g.width/2, 2*g.height/3)
}

// drawPauseMenu handles the drawing logic for the pause menu
func (g *Game) drawPauseMenu() {
	g.renderer.DrawText("PAUSED", g.width/2, g.height/3)
	g.renderer.DrawText("Press P to resume", g.width/2, g.height/2)
}

// Update updates the game state
func (g *Game) Update(dt float64) error {
	g.logger.Debug("Updating game state", "state", g.state, "dt", dt)
	switch g.state {
	case MainMenu:
		g.updateMainMenu(dt)
	case Playing:
		g.updatePlaying(dt)
	case GameOver:
		g.updateGameOver(dt)
	case PauseMenu:
		g.updatePauseMenu(dt)
	}
	return nil
}

// updateMainMenu handles updates for the main menu state
func (g *Game) updateMainMenu(dt float64) {
	// Add any animations or effects for the main menu here
	// For example, you could have a blinking "Press ENTER to start" text
	g.blinkTimer += dt
	if g.blinkTimer >= g.blinkInterval {
		g.blinkTimer = 0
		g.showPressEnter = !g.showPressEnter
	}
}

// updatePlaying handles the update logic for the playing state
func (g *Game) updatePlaying(dt float64) {
	g.logger.Debug("Updating playing state")
	g.movePlayer(dt)
	g.moveAliens(dt)
	g.moveBullets(dt)
	g.checkCollisions()
	g.checkGameOver()
}

// updateGameOver handles updates for the game over state
func (g *Game) updateGameOver(dt float64) {
	// Add any animations or effects for the game over screen here
	// For example, you could have a score counter that counts up to the final score
	if g.displayScore < g.score {
		g.displayScore += int(float64(g.score-g.displayScore) * dt * 2)
		if g.displayScore > g.score {
			g.displayScore = g.score
		}
	}
}

// updatePauseMenu handles updates for the pause menu state
func (g *Game) updatePauseMenu(dt float64) {
	// Add any animations or effects for the pause menu here
	// For example, you could have a pulsing "PAUSED" text
	g.pausePulseTimer += dt
	g.pausePulseScale = 1 + 0.1*float64(g.pausePulseTimer)
	if g.pausePulseTimer >= 1 {
		g.pausePulseTimer = 0
	}
}

// HandleInput processes user input
func (g *Game) HandleInput(input core.InputEvent) error {
	g.logger.Debug("Handling input", "type", input.Type, "key", input.Key)
	switch g.state {
	case MainMenu:
		if input.Type == core.KeyPress && input.Key == core.KeyEnter {
			g.startNewGame()
		}
	case Playing:
		return g.handlePlayingInput(input)
	case GameOver:
		if input.Type == core.KeyPress && input.Key == core.KeyEnter {
			g.startNewGame()
		}
	case PauseMenu:
		if input.Type == core.KeyPress && input.Key == core.KeyBackspace {
			g.state = Playing
			g.logger.Info("Game resumed")
		}
	}
	return nil
}

// handlePlayingInput processes user input during the playing state
func (g *Game) handlePlayingInput(input core.InputEvent) error {
	if input.Type == core.KeyPress {
		switch input.Key {
		case core.KeySpace:
			g.fireBullet()
		case core.KeyBackspace:
			g.state = PauseMenu
			g.logger.Info("Game paused")
		}
	}
	return nil
}

// spawnAliens creates a new wave of aliens
func (g *Game) spawnAliens() {
	const (
		rows         = 5
		aliensPerRow = 11
		xPadding     = 50
		yPadding     = 50
	)

	for row := 0; row < rows; row++ {
		for col := 0; col < aliensPerRow; col++ {
			alienType := AlienType(row / 2)
			alien := &Alien{
				Entity: Entity{
					Position: Vector2D{
						X: float64(col*xPadding) + xPadding,
						Y: float64(row*yPadding) + yPadding,
					},
					Speed: AlienSpeed,
				},
				Type:   alienType,
				Points: (3 - int(alienType)) * 10,
			}
			g.aliens = append(g.aliens, alien)
		}
	}
	g.logger.Info("Spawned new wave of aliens", "level", g.level)
}

// createBarriers initializes the defensive barriers
func (g *Game) createBarriers() {
	const (
		barrierCount = 4
		barrierWidth = 60
		barrierY     = 150 // Distance from bottom
	)

	for i := 0; i < barrierCount; i++ {
		barrier := &Barrier{
			Entity: Entity{
				Position: Vector2D{
					X: float64(i+1)*(float64(g.width)/(barrierCount+1)) - barrierWidth/2,
					Y: float64(g.height) - barrierY,
				},
			},
			Health: 4,
		}
		g.barriers = append(g.barriers, barrier)
	}
	g.logger.Info("Created defensive barriers")
}

// moveAliens updates the positions of all aliens
func (g *Game) moveAliens(dt float64) {
	moveDown := false
	for _, alien := range g.aliens {
		alien.Position.X += alien.Speed * dt
		if alien.Position.X <= 0 || alien.Position.X >= float64(g.width) {
			moveDown = true
		}
	}

	if moveDown {
		g.logger.Debug("Aliens moving down")
		for _, alien := range g.aliens {
			alien.Speed = -alien.Speed
			alien.Position.Y += 10
		}
	}
}

// moveBullets updates the positions of all bullets
func (g *Game) moveBullets(dt float64) {
	initialCount := len(g.bullets)
	for i := len(g.bullets) - 1; i >= 0; i-- {
		bullet := g.bullets[i]
		bullet.Position.Y -= bullet.Speed * dt

		// Remove bullets that are off-screen
		if bullet.Position.Y < 0 {
			g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
		}
	}
	removedCount := initialCount - len(g.bullets)
	if removedCount > 0 {
		g.logger.Debug("Removed off-screen bullets", "count", removedCount)
	}
}

// checkCollisions detects and handles collisions between game objects
func (g *Game) checkCollisions() {
	alienHitCount := 0
	barrierHitCount := 0

	// Check bullet collisions with aliens
	for i := len(g.bullets) - 1; i >= 0; i-- {
		bullet := g.bullets[i]
		for j := len(g.aliens) - 1; j >= 0; j-- {
			alien := g.aliens[j]
			if checkCollision(bullet.Position, alien.Position, 5, 30) {
				g.updateScore(alien.Points)
				g.aliens = append(g.aliens[:j], g.aliens[j+1:]...)
				g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
				alienHitCount++
				break
			}
		}
	}

	// Check bullet collisions with barriers
	for i := len(g.bullets) - 1; i >= 0; i-- {
		bullet := g.bullets[i]
		for _, barrier := range g.barriers {
			if checkCollision(bullet.Position, barrier.Position, 5, 60) {
				barrier.Health--
				g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
				barrierHitCount++
				break
			}
		}
	}

	// Remove destroyed barriers
	initialBarrierCount := len(g.barriers)
	for i := len(g.barriers) - 1; i >= 0; i-- {
		if g.barriers[i].Health <= 0 {
			g.barriers = append(g.barriers[:i], g.barriers[i+1:]...)
		}
	}
	destroyedBarrierCount := initialBarrierCount - len(g.barriers)

	if alienHitCount > 0 || barrierHitCount > 0 || destroyedBarrierCount > 0 {
		g.logger.Info("Collision results",
			"aliensHit", alienHitCount,
			"barriersHit", barrierHitCount,
			"barriersDestroyed", destroyedBarrierCount)
	}
}

// checkGameOver determines if the game should end
func (g *Game) checkGameOver() {
	if len(g.aliens) == 0 {
		g.level++
		g.logger.Info("Level completed", "newLevel", g.level)
		g.spawnAliens()
		return
	}

	for _, alien := range g.aliens {
		if alien.Position.Y >= g.player.Position.Y {
			g.state = GameOver
			g.logger.Info("Game over: Aliens reached the bottom")
			return
		}
	}

	if g.player.lives <= 0 {
		g.state = GameOver
		g.logger.Info("Game over: Player out of lives")
	}
}

// Helper functions

func checkCollision(pos1, pos2 Vector2D, size1, size2 float64) bool {
	return abs(pos1.X-pos2.X) < (size1+size2)/2 && abs(pos1.Y-pos2.Y) < (size1+size2)/2
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// movePlayer updates the player's position based on input
func (g *Game) movePlayer(dt float64) {
	initialX := g.player.Position.X
	for _, event := range g.inputHandler.PollEvents() {
		if event.Type == core.KeyPress {
			switch event.Key {
			case core.KeyLeft:
				g.player.Position.X -= g.player.Speed * dt
			case core.KeyRight:
				g.player.Position.X += g.player.Speed * dt
			}
		}
	}

	// Clamp player position to screen bounds
	g.player.Position.X = clamp(g.player.Position.X, PlayerSize/2, float64(g.width)-PlayerSize/2)

	if g.player.Position.X != initialX {
		g.logger.Debug("Player moved", "newX", g.player.Position.X)
	}
}

// fireBullet creates a new bullet from the player's position
func (g *Game) fireBullet() {
	bullet := &Bullet{
		Entity: Entity{
			Position: Vector2D{X: g.player.Position.X, Y: g.player.Position.Y - 10},
			Speed:    BulletSpeed,
		},
		Damage: 1,
	}
	g.bullets = append(g.bullets, bullet)
	g.logger.Info("Player fired a bullet")
}

// updateScore increases the player's score
func (g *Game) updateScore(points int) {
	g.score += points
	g.logger.Info("Score updated", "score", g.score)
}

// startNewGame initializes a new game
func (g *Game) startNewGame() {
	g.logger.Info("Starting new game")
	g.score = 0
	g.level = 1
	g.player.lives = 3
	g.player.Position.X = float64(g.width) / 2
	g.aliens = nil
	g.bullets = nil
	g.barriers = nil
	g.spawnAliens()
	g.createBarriers()
	g.state = Playing
	g.displayScore = 0
	g.pausePulseTimer = 0
	g.pausePulseScale = 1
}