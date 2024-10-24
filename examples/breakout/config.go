package breakout

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

	// Paddle config
	PaddleWidth     float64
	PaddleHeight    float64
	PaddleSpeed     float64
	PaddleYPosition float64

	// Ball config
	BallSize            float64
	BallSpeed           float64
	BallVelocityX      float64
	BallVelocityY      float64
	BallMinXVelocity   float64
	BallMinYVelocity   float64

	// Brick config
	BrickWidth   float64
	BrickHeight  float64
	BrickRows    int
	BrickSpacing float64
	BrickStartY  float64

	// Game settings
	InitialLives    int
	MaxNameLength   int
	BlinkInterval   float64
	LeaderboardSize int
}

func NewConfig(workDir string) (*Config, error) {
	cfg := &Config{
		BaseConfig:      config.NewBaseConfig(workDir, "Breakout"),
		TitleOffset:     0.1,
		LineSpacing:     2,
		ScoreOffset:     1.0 / 6,
		ControlsOffset:  2.0 / 8,

		PaddleWidth:     10.0,
		PaddleHeight:    1.0,
		PaddleSpeed:     1.0,
		PaddleYPosition: 2.0,

		BallSize:          1.0,
		BallSpeed:         10.0,
		BallVelocityX:     10.0,
		BallVelocityY:     -10.0,
		BallMinXVelocity:  2.0,
		BallMinYVelocity:  5.0,

		BrickWidth:   8.0,
		BrickHeight:  1.0,
		BrickRows:    4,
		BrickSpacing: 2.0,
		BrickStartY:  3.0,

		InitialLives:    3,
		MaxNameLength:   20,
		BlinkInterval:   0.5,
		LeaderboardSize: 5,
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
