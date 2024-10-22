package gameoflife

import (
	"os"

	"github.com/kuhree/gg/internal/engine/config"
)

// Config holds all the game configuration values
type Config struct {
	config.BaseConfig

	BaseNeighboars int
	BaseChance     float64
	BaseSize       int
	BaseMoveSpeed  float64
	BaseRadius     float64
	SeedBuffer     float64

	StabilityThreshold int
	StabilityChance    float64
}

func NewConfig(workDir string) (*Config, error) {
	cfg := &Config{
		BaseConfig:         config.NewBaseConfig(workDir, "Conway's Game of Life"),
		BaseNeighboars:     8,
		BaseChance:         0.20, // 20% chance to be alive initially
		BaseSize:           1,
		BaseMoveSpeed:      1.0,
		BaseRadius:         1.0,
		SeedBuffer:         1.0,
		StabilityThreshold: 500,
		StabilityChance:    0.90,
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
