package sorts

import (
	"github.com/kuhree/gg/internal/engine/config"
	"os"
)

type Config struct {
	config.BaseConfig

	// Scene layout
	TitleOffset    float64
	LineSpacing    int
	ScoreOffset    float64
	ControlsOffset float64

	// Visualization settings
	ArraySize      int
	UpdateInterval float64
	BarWidth      float64
	MaxValue      int
}

func NewConfig(workDir string) (*Config, error) {
	cfg := &Config{
		BaseConfig:     config.NewBaseConfig(workDir, "Sorting Visualizer"),
		TitleOffset:    0.1,
		LineSpacing:    2,
		ScoreOffset:    1.0 / 6,
		ControlsOffset: 2.0 / 8,

		ArraySize:      100,
		UpdateInterval: 0.1,
		BarWidth:      1,
		MaxValue:      100,
	}

	err := config.LoadConfig(cfg)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
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
