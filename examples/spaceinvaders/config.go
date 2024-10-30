package space_invaders

import (
	"os"

	"github.com/kuhree/gg/internal/engine/config"
)

// Config holds all the game configuration values
type Config struct {
	config.BaseConfig

	TargetFPS      int
	PlayerYOffset  int
	BarrierYOffset int
	AlienYOffset   int

	BaseLevel     int
	BaseLevelStep int
	BaseScore     int
	BaseLives     int

	BaseDifficulty           float64
	BaseDifficultyMultiplier float64

	BaseCollectibleDuration      float64
	BaseCollectableSpawnInterval float64
	BaseMaxCollectables          int
	BaseCollectableSpeed         float64

	BasePlayerSize   float64
	BasePlayerSpeed  float64
	BasePlayerHealth float64

	BaseAliensCount int
	BaseAlienSize   float64
	BaseAlienSpeed  float64
	BaseAlienHealth float64

	BaseProjectileSize   float64
	BaseProjectileSpeed  float64
	BaseProjectileHealth float64

	BaseBarrierSize             float64
	BaseBarrierCount            int
	BaseBarrierMinimum          int
	BaseBarrierHealth           float64
	BaseBarrierRegenerationRate float64

	BaseShootInterval     float64
	MinShootInterval      float64
	ShootIntervalVariance float64
	BaseShootChance       float64
	CooldownMultiplier    float64
	IntervalRandomFactor  float64
}

func NewConfig(workDir string) (*Config, error) {
	cfg := &Config{
		BaseConfig:     config.NewBaseConfig(workDir, "Space Invaders"),
		PlayerYOffset:  3,
		BarrierYOffset: 7,
		AlienYOffset:   3,

		BaseScore:                1,
		BaseLives:                3,
		BaseLevel:                1,
		BaseLevelStep:            1,
		BaseDifficulty:           1.0,
		BaseDifficultyMultiplier: 0.2,

		BaseCollectibleDuration:      10.0,
		BaseCollectableSpawnInterval: 10.0,
		BaseMaxCollectables:          3,
		BaseCollectableSpeed:         1,

		BasePlayerSize:   2.0,
		BasePlayerSpeed:  1.0,
		BasePlayerHealth: 10.0,

		BaseAliensCount: 1,
		BaseAlienSize:   2.0,
		BaseAlienSpeed:  1.0,
		BaseAlienHealth: 20.0,

		BaseProjectileSize:   1.0,
		BaseProjectileSpeed:  30.0,
		BaseProjectileHealth: 1.0,

		BaseBarrierCount:            7,
		BaseBarrierSize:             3.0,
		BaseBarrierHealth:           50.0,
		BaseBarrierRegenerationRate: 2,
		BaseBarrierMinimum:          2,

		BaseShootInterval:     8.0,
		MinShootInterval:      3.0,
		ShootIntervalVariance: 20.0,
		BaseShootChance:       0.4,
		CooldownMultiplier:    1.5,
		IntervalRandomFactor:  0.5,
	}

	err := config.LoadConfig(cfg)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		// If the file doesn't exist, save the default config
		err = config.SaveConfig(cfg)
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func (c *Config) Save() error {
	return config.SaveConfig(c)
}

func (c *Config) Load() error {
	return config.LoadConfig(c)
}
