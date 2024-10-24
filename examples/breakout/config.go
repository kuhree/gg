package breakout

import (
	"os"

	"github.com/kuhree/gg/internal/engine/config"
)

// Config holds all the game configuration values
type Config struct {
	config.BaseConfig

}

func NewConfig(workDir string) (*Config, error) {
	cfg := &Config{
		BaseConfig:         config.NewBaseConfig(workDir, "Breakout"),
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
