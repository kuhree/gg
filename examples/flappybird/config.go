package flappybird

import (
	"os"

	"github.com/kuhree/gg/internal/engine/config"
)

// Config holds all the game configuration values
type Config struct {
	config.BaseConfig

	// Scene layout
	TitleOffset    float64
	LineSpacing    int
	ScoreOffset    float64
	ControlsOffset float64

	// Game settings
	InitialLives    int
	MaxNameLength   int
	BlinkInterval   float64
	LeaderboardSize int

	// Bird physics
	BirdGravity   float64
	BirdJumpForce float64
	
	// Pipe settings
	PipeSpeed     float64
	PipeGap       float64
	PipeSpacing   float64
	MinPipeHeight float64
	PipeWidth     float64
}

func NewConfig(workDir string) (*Config, error) {
	cfg := &Config{
		BaseConfig:      config.NewBaseConfig(workDir, "Flappy Bird"),
		TitleOffset:     0.1,
		LineSpacing:     2,
		ScoreOffset:     1.0 / 6,
		ControlsOffset:  2.0 / 8,

		InitialLives:    3,
		MaxNameLength:   20,
		BlinkInterval:   0.5,
		LeaderboardSize: 5,

		BirdGravity:   20.0,
		BirdJumpForce: -10.0,
		
		PipeSpeed:     15.0,
		PipeGap:       8.0,
		PipeSpacing:   20.0,
		MinPipeHeight: 3.0,
		PipeWidth:     2.0,
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
