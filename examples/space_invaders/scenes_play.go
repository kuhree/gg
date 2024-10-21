package space_invaders

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/ojrac/opensimplex-go"

	"github.com/kuhree/gg/internal/engine/render"
)

const (
	BasicAlien AlienType = iota
	FastAlien
	ToughAlien
	ShooterAlien
	BossAlien
)

type AlienAttributes struct {
	Type        AlienType
	Size        int
	Health      float64
	Speed       float64
	AttackPower float64
	ShootChance float64
}

var alienTypes = map[AlienType]AlienAttributes{
	BasicAlien: {
		Type:        BasicAlien,
		Size:        2,
		Health:      10,
		Speed:       2,
		AttackPower: 1,
		ShootChance: 0.10,
	},
	FastAlien: {
		Type:        FastAlien,
		Size:        3,
		Health:      5,
		Speed:       5,
		AttackPower: 1,
		ShootChance: 0.05,
	},
	ToughAlien: {
		Type:        ToughAlien,
		Size:        4,
		Health:      50,
		Speed:       1,
		AttackPower: 2,
		ShootChance: 0.005,
	},
	ShooterAlien: {
		Type:        ShooterAlien,
		Size:        3,
		Health:      8,
		Speed:       7,
		AttackPower: 1.5,
		ShootChance: 0.50,
	},
	BossAlien: {
		Type:        BossAlien,
		Size:        5,
		Health:      100,
		Speed:       3,
		AttackPower: 3,
		ShootChance: 0.05,
	},
}

// updateAliens updates the positions of all aliens
func (s *PlayingScene) updateAliens(dt float64) {
	width, height := s.Renderer.Size()

	for _, alien := range s.Aliens {
		// Apply smoother movement based on alien type
		switch alien.AlienType {
		case FastAlien:
			alien.Position.X += alien.Speed.X * dt
			alien.Position.Y += math.Sin(alien.Position.X*0.1) * dt * 10
		case ToughAlien:
			alien.Position.X += alien.Speed.X * dt * 0.8
			alien.Position.Y += math.Cos(alien.Position.X*0.05) * dt * 5
		case ShooterAlien:
			alien.Position.X += alien.Speed.X * dt
			alien.Position.Y += math.Sin(alien.Position.X*0.2) * dt * 5
		case BossAlien:
			alien.Position.X += alien.Speed.X * dt
			alien.Position.Y += math.Cos(alien.Position.X*0.05) * dt * 5
		default:
			// Add a slight vertical movement to basic aliens
			alien.Position.X += alien.Speed.X * dt
			alien.Position.Y += math.Sin(alien.Position.X*0.1) * dt * 2
		}

		// Smooth boundary checks
		if alien.Position.X-alien.Width/2 <= 0 || alien.Position.X+alien.Width/2 >= float64(width) {
			// Gradually change direction
			alien.Speed.X = -alien.Speed.X * 0.9
		}

		if alien.Position.Y-alien.Height/2 <= 0 || alien.Position.Y+alien.Height/2 >= float64(height) {
			// Gradually change direction
			alien.Speed.Y = -alien.Speed.Y * 0.9
		}

		// Ensure alien stays within bounds
		alien.Position.X = clamp(alien.Position.X, alien.Width/2, float64(width)-alien.Width/2)
		alien.Position.Y = clamp(alien.Position.Y, alien.Height/2, float64(height)-alien.Height/2)

	}
}

// updateAlienFiringSquad handles aliens "shooting"
func (s *PlayingScene) updateAlienFiringSquad(dt float64) {
	for _, alien := range s.Aliens {
		alien.shootCooldown -= dt
		if alien.shootCooldown <= 0 {
			if rand.Float64() < alien.shootChance {
				cooldownRandomFactor := rand.Float64() * s.Config.IntervalRandomFactor
				alien.shootCooldown = alien.shootInterval * (1 + cooldownRandomFactor)
				s.Shoot(&alien.GameObject)
			}
			alien.shootCooldown = alien.shootInterval
		}
	}
}

// updateProjectiles updates the positions of all projectiles
func (s *PlayingScene) updateProjectiles(dt float64) {
	for i := len(s.Projectiles) - 1; i >= 0; i-- {
		projectile := s.Projectiles[i]
		projectile.Position.X += projectile.Speed.X * dt
		projectile.Position.Y += projectile.Speed.Y * dt

		// Remove projectiles that are off-screen
		if projectile.Position.Y < 0 {
			s.Projectiles[i].Health = 0
			s.Projectiles = append(s.Projectiles[:i], s.Projectiles[i+1:]...)
		}
	}
}

// movePlayer updates the player's position based on the given direction
func (s *PlayingScene) movePlayer(dx, dy int) {
	newX := s.Player.Position.X + float64(dx)*s.Player.Speed.X
	newY := s.Player.Position.Y + float64(dy)*s.Player.Speed.Y

	// Clamp the player's position to stay within the game boundaries
	width, height := s.Renderer.Size()
	newX = clamp(newX, float64(s.Player.Width)/2, float64(width)-float64(s.Player.Width)/2)
	newY = clamp(newY, float64(s.Player.Height)/2, float64(height)-float64(s.Player.Height)/2)

	s.Player.Position.X = newX
	s.Player.Position.Y = newY

	s.Logger.Debug("Player moved", "newX", newX, "newY", newY, "dx", dx, "dy", dy)
}

// Shoot creates a new projectile from the given position
// / recoil ??
func (s *PlayingScene) Shoot(source *GameObject) {
	isFromPlayer := source == &s.Player.GameObject
	position := source.Position
	attack := source.Attack

	from := "Alien"
	speed := Vector2D{X: 0, Y: s.Config.BaseProjectileSpeed / 2}
	newY := (position.Y + float64(source.Height)/2) + 0.5
	if isFromPlayer {
		from = "Player"
		speed = Vector2D{X: 0, Y: -s.Config.BaseProjectileSpeed}
		newY = (position.Y - float64(source.Width)/2) - 0.5
	}

	projectile := &Projectile{
		GameObject: GameObject{
			Position:  Vector2D{X: position.X, Y: newY},
			Speed:     speed,
			Health:    s.Config.BaseProjectileHealth,
			MaxHealth: s.Config.BaseProjectileHealth,
			Attack:    attack,
			Width:     s.Config.BaseProjectileSize,
			Height:    s.Config.BaseProjectileSize,
		},

		Source: source,
	}

	s.Projectiles = append(s.Projectiles, projectile)
	s.Logger.Info(fmt.Sprintf("%s shot a projectile", from), "speed", projectile.Speed)
}

// updateCollisions detects and handles collisions between game objects
// collisions are something else
func (s *PlayingScene) updateCollisions() {
	player := s.Player

	for i := len(s.Aliens) - 1; i >= 0; i-- {
		alien := s.Aliens[i]

		// check alien/player collisions
		if alien.Health >= 0 && s.Collides(&player.GameObject, &alien.GameObject) {
			if alien.Health >= 0 {
				s.Player.Health -= alien.Health
			}

			s.Aliens[i].Health = 0 // kill it, don't wanna "bump" into it 1n times and die
		}

		// check alien/barrier collisions
		for j, barrier := range s.Barriers {
			if barrier.Health > 0 && s.Collides(&alien.GameObject, &barrier.GameObject) {
				s.Barriers[j].Health -= alien.Attack
				s.Aliens[i].Health = 0
			}
		}
	}

	// Check all projectile collisions
	for i := len(s.Projectiles) - 1; i >= 0; i-- {
		projectile := s.Projectiles[i]
		isFromPlayer := projectile.Source == &player.GameObject

		// projectile/player
		if !isFromPlayer && projectile.Health >= 0 && s.Collides(&projectile.GameObject, &player.GameObject) {
			s.Player.Health -= projectile.Attack
			s.Projectiles[i].Health = 0 // kill it w fire NOW
		}

		// projectile/alien
		for j := len(s.Aliens) - 1; j >= 0; j-- {
			alien := s.Aliens[j]
			if isFromPlayer && alien.Health >= 0 && s.Collides(&projectile.GameObject, &alien.GameObject) {
				if alien.Health >= 0 {
					s.Projectiles[i].Health -= alien.Health
				}

				score := alien.MaxHealth
				s.Aliens[j].Health -= projectile.Attack
				if alien.Health <= 0 {
					s.increaseScore(int(score))
				}
			}
		}

		// projectile/barrier
		for j := len(s.Barriers) - 1; j >= 0; j-- {
			barrier := s.Barriers[j]
			if !isFromPlayer && barrier.Health >= 0 && s.Collides(&projectile.GameObject, &barrier.GameObject) {
				s.Barriers[j].Health -= projectile.Attack
				s.Projectiles[i].Health = 0 // kill it, issa barrier
			}
		}

		// projectile/projectile
		for j := len(s.Projectiles) - 1; j >= 0; j-- {
			proj := s.Projectiles[j]
			isFromPlayerInner := &player.GameObject == proj.Source

			if isFromPlayer && !isFromPlayerInner && proj.Health >= 0 && s.Collides(&projectile.GameObject, &proj.GameObject) {
				s.Projectiles[j].Health -= projectile.Attack
				s.Projectiles[i].Health = proj.Attack
			}
		}
	}
}

// Collides checks if two GameObjects are colliding
func (s *PlayingScene) Collides(obj1, obj2 *GameObject) bool {
	// Calculate the edges of each object
	left1 := obj1.Position.X - obj1.Width/2
	right1 := obj1.Position.X + obj1.Width/2
	top1 := obj1.Position.Y - obj1.Height/2
	bottom1 := obj1.Position.Y + obj1.Height/2

	left2 := obj2.Position.X - obj2.Width/2
	right2 := obj2.Position.X + obj2.Width/2
	top2 := obj2.Position.Y - obj2.Height/2
	bottom2 := obj2.Position.Y + obj2.Height/2

	// Check for overlap
	if left1 < right2 && right1 > left2 && top1 < bottom2 && bottom1 > top2 {
		return true
	}

	return false
}

func (s *PlayingScene) killTheDead() {
	_, height := s.Renderer.Size()
	for i := len(s.Aliens) - 1; i >= 0; i-- {
		alien := s.Aliens[i]
		if alien.Health <= 0 {
			s.Aliens = append(s.Aliens[:i], s.Aliens[i+1:]...)
		} else if alien.AlienType == BasicAlien && alien.Health > 0 && alien.Position.Y+float64(alien.Height)/2 >= float64(height) {
			s.Logger.Warn("Alien reached the bottom")
			s.Aliens = append(s.Aliens[:i], s.Aliens[i+1:]...)
			s.Player.Health = 0
		}
	}

	for i := len(s.Barriers) - 1; i >= 0; i-- {
		barrier := s.Barriers[i]
		if barrier.Health <= 0 {
			s.Barriers = append(s.Barriers[:i], s.Barriers[i+1:]...)
		}
	}

	for i := len(s.Projectiles) - 1; i >= 0; i-- {
		projectile := s.Projectiles[i]
		if projectile.Health <= 0 {
			s.Projectiles = append(s.Projectiles[:i], s.Projectiles[i+1:]...)
		}
	}
}

// updateGameState determines if the game should end
func (s *PlayingScene) updateGameState() {
	width, height := s.Renderer.Size()

	if len(s.Aliens) == 0 {
		s.increaseLevel()
		return
	}

	if s.Player.Health <= 0 {
		s.Logger.Info("Player is out of health. Losing life...")
		s.Player.Lives--

		if s.Player.Lives > 0 {
			s.Player.Health = s.Config.BasePlayerHealth
			s.Player.MaxHealth = s.Config.BasePlayerHealth
			s.Player.Position = Vector2D{X: float64(width) / 2, Y: float64(height - s.Config.PlayerYOffset)}
			return
		}
	}

	if s.Player.Lives <= 0 {
		s.Logger.Info("Player out of lives. Game over...")
		s.Scenes.ChangeScene(GameOverSceneID)
		return
	}
}

// startWave configures the game state for the current level
func (s *PlayingScene) startWave() {
	// Reset game entities
	s.Aliens = nil
	s.Projectiles = nil

	difficultyMultiplier := s.difficulty()
	s.setupLevelPlayer(difficultyMultiplier)
	s.setupLevelAliens(difficultyMultiplier)
	s.setupLevelBarriers(difficultyMultiplier)

	s.Logger.Info("Level setup complete",
		"level", s.CurrentLevel,
		"aliens", len(s.Aliens),
		"barriers", len(s.Barriers),
	)
}

func (s *PlayingScene) increaseScore(delta int) int {
	s.Logger.Info("Increasing score!", "score", s.Score, "delta", delta, "newScore", s.Score+delta)
	s.Score += delta
	return s.Score
}

func (s *PlayingScene) increaseLevel() {
	s.Logger.Info("Level cleared! Advancing...", "newLevel", s.CurrentLevel+s.Config.BaseLevelStep)
	s.CurrentLevel += s.Config.BaseLevelStep

	s.startWave()
}

func (s *PlayingScene) difficulty() float64 {
	return s.Config.BaseDifficulty + float64(s.CurrentLevel)*s.Config.BaseDifficultyMultiplier
}

func (s *PlayingScene) setupLevelPlayer(difficultyMultiplier float64) {
	width, height := s.Renderer.Size()

	s.Player.Position = Vector2D{X: float64(width) / 2, Y: float64(height - s.Config.PlayerYOffset)}
	s.Player.Attack = s.Config.BasePlayerAttack * max(1.1, difficultyMultiplier*0.66)
}

func (s *PlayingScene) setupLevelAliens(difficultyMultiplier float64) {
	width, height := s.Renderer.Size()

	// Calculate number of aliens
	alienCount := s.CurrentLevel + s.Config.BaseAliensCount*2
	if s.CurrentLevel <= 3 {
		alienCount = s.CurrentLevel
	}

	// Generate aliens first
	aliens := s.generateAliens(alienCount, difficultyMultiplier)

	// Generate alien positions
	positions := s.generateAlienPositions(aliens, width, height)

	// Assign positions to aliens
	for i, pos := range positions {
		alien := aliens[i]
		alien.Position = pos

		s.Aliens = append(s.Aliens, alien)
	}
}

func (s *PlayingScene) generateAliens(count int, difficultyMultiplier float64) []*Alien {
	aliens := make([]*Alien, 0, count)

	adjustedShootInterval := max(s.Config.BaseShootInterval/difficultyMultiplier, s.Config.MinShootInterval)

	for i := 0; i < count; i++ {
		alienType := BasicAlien
		if i == 0 && s.CurrentLevel%10 == 0 {
			alienType = BossAlien
		} else if i <= 1 {
			alienType = BasicAlien
		} else if i%7 == 0 {
			alienType = FastAlien
		} else if i%5 == 0 {
			alienType = ToughAlien
		} else if i%3 == 0 {
			alienType = ShooterAlien
		}

		attributes := alienTypes[alienType]
		health := attributes.Health * difficultyMultiplier
		alien := &Alien{
			GameObject: GameObject{
				Speed:     Vector2D{X: attributes.Speed * difficultyMultiplier, Y: 0},
				Width:     float64(attributes.Size),
				Height:    float64(attributes.Size),
				Health:    health,
				MaxHealth: health,
				Attack:    attributes.AttackPower * difficultyMultiplier,
			},
			AlienType:     alienType,
			shootInterval: adjustedShootInterval,
			shootCooldown: rand.Float64() * adjustedShootInterval * s.Config.CooldownMultiplier,
			shootChance:   attributes.ShootChance * difficultyMultiplier,
		}
		aliens = append(aliens, alien)
	}

	return aliens
}

func (s *PlayingScene) generateAlienPositions(aliens []*Alien, width, height int) []Vector2D {
	noise := opensimplex.NewNormalized(int64(s.CurrentLevel))
	positions := make([]Vector2D, 0, len(aliens))

	topMargin := s.Config.AlienYOffset
	bottomMargin := s.Config.BarrierYOffset
	sideMargin := int(s.Config.BaseAlienSize / 2)

	spawnWidth := width - 2*sideMargin
	spawnHeight := height - topMargin - bottomMargin

	centerX := float64(width) / 2
	centerY := float64(topMargin + spawnHeight/2)

	for i := 0; i < len(aliens)*10; i++ { // Increase iterations to ensure we get enough valid positions
		angle := noise.Eval2(float64(i)*0.1, 0) * 2 * math.Pi
		distance := math.Sqrt(noise.Eval2(0, float64(i)*0.1)) * 0.8 // Use sqrt for less aggressive centering, limit to 80% of max distance

		relX := math.Cos(angle) * distance * float64(spawnWidth/2)
		relY := math.Sin(angle) * distance * float64(spawnHeight/2)

		x := centerX + relX
		y := centerY + relY

		// Check if the position is valid
		if x >= float64(sideMargin) && x <= float64(width-sideMargin) &&
			y >= float64(topMargin) && y <= float64(height-bottomMargin) {

			// Check for overlap with existing positions
			overlap := false
			for j, pos := range positions {
				minDistance := (aliens[len(positions)].Width + aliens[j].Width) / 2
				if math.Hypot(x-pos.X, y-pos.Y) < minDistance {
					overlap = true
					break
				}
			}

			if !overlap {
				positions = append(positions, Vector2D{X: x, Y: y})
			}
		}

		if len(positions) >= len(aliens) {
			break
		}
	}

	return positions
}

func (s *PlayingScene) setupLevelBarriers(difficultyMultiplier float64) {
	if len(s.Barriers) <= 0 {
		width, height := s.Renderer.Size()
		s.BarriersCountLast += 1

		barrierCount := max(s.Config.BaseBarrierCount-int(difficultyMultiplier), 1)
		for i := 0; i < barrierCount; i++ {
			health := s.Config.BaseBarrierHealth
			barrier := &Barrier{
				GameObject: GameObject{
					Position: Vector2D{
						X: float64(i+1) * (float64(width) / (float64(barrierCount) + 1)),
						Y: float64(height - s.Config.BarrierYOffset),
					},
					Speed:     Vector2D{},
					Health:    health,
					MaxHealth: health,
					Attack:    s.Config.BaseBarrierAttack,
					Width:     s.Config.BaseBarrierSize * 2,
					Height:    s.Config.BaseBarrierSize,
				},
			}

			s.Barriers = append(s.Barriers, barrier)
		}
	}
}

func (s *PlayingScene) getAlienInfo(alien *Alien) (rune, render.Color) {
	ratio := alien.Health / float64(alien.MaxHealth)

	alienConfigs := map[AlienType]struct {
		Char  rune
		Color render.Color
	}{
		FastAlien:    {'~', render.ColorCyan},
		ToughAlien:   {'#', render.ColorGreen},
		BossAlien:    {'T', render.ColorRed},
		ShooterAlien: {'+', render.ColorBlue},
		BasicAlien:   {render.FullBlock, render.ColorWhite},
	}

	conf, ok := alienConfigs[alien.AlienType]
	if !ok {
		conf = alienConfigs[BasicAlien]
	}

	// Adjust appearance based on health
	switch {
	case ratio > 1.5:
		// conf.Color = render.ColorBlue
	case ratio > 0.75:
		// No change, use base appearance
	case ratio > 0.5:
		conf.Char = render.MediumShade
	case ratio > 0.25:
		conf.Char = render.LightShade
		// conf.Color = render.ColorYellow
	case ratio > 0:
		conf.Char = render.WhiteSquare
		// conf.Color = render.ColorRed
	default:
		conf.Char = render.BlackSquare
		// conf.Color = render.ColorRed
	}

	return conf.Char, conf.Color
}

func (s *PlayingScene) getBarrierInfo(health float64, maxHealth float64) (rune, render.Color) {
	ratio := health / float64(maxHealth)
	switch {
	case ratio < 1.0:
		return s.getHealthInfo(health, maxHealth)
	default:
		return render.FullBlock, render.ColorGreen
	}
}

func (s *PlayingScene) getHealthInfo(health float64, maxHealth float64) (rune, render.Color) {
	ratio := health / float64(maxHealth)
	switch {
	case ratio > 1.0:
		return render.FullBlock, render.ColorBlue
	case ratio == 1.0:
		return render.FullBlock, render.ColorWhite
	case ratio >= 1.0:
		return render.DarkShade, render.ColorWhite
	case ratio >= 0.75:
		return render.LightShade, render.ColorYellow
	case ratio >= 0.5:
		return render.MediumShade, render.ColorYellow
	case ratio >= 0.25:
		return render.LightShade, render.ColorRed
	default:
		return render.FullBlock, render.ColorWhite
	}
}

func (s *PlayingScene) getProjectileInfo(proj *Projectile) (rune, render.Color) {
	var char rune
	var color render.Color

	isFromPlayer := proj.Source == &s.Player.GameObject

	attackRatio := proj.Attack / s.Player.Health
	if isFromPlayer {
		attackRatio = proj.Attack / s.Config.BasePlayerAttack
	}

	switch {
	case attackRatio <= 1:
		char, color = '.', render.ColorWhite // Very weak attack
		if isFromPlayer {
			char = 's'
		}
	case attackRatio <= 1.25:
		char, color = '|', render.ColorBlue // Weak attack
	case attackRatio <= 1.75:
		char, color = '+', render.ColorCyan // Moderate attack
	case attackRatio < 2.5:
		char, color = '*', render.ColorGreen // Strong attack
	case attackRatio < 3:
		char, color = render.WhiteTriangle, render.ColorRed // Very strong attack
	case attackRatio < 5:
		char, color = render.BlackTriangle, render.ColorRed // Extremely powerful attack
	default:
		char, color = render.FullBlock, render.ColorRed // Extremely powerful attack
	}

	ratio := proj.Health / float64(proj.MaxHealth)
	switch {
	case ratio <= 0.25:
		char, color = render.LightShade, render.ColorRed
	case ratio <= 0.5:
		char, color = render.MediumShade, render.ColorYellow
	case ratio <= 0.75:
		char, color = render.LightShade, render.ColorYellow
	}

	return char, color
}
