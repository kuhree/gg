package space_invaders

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
)

// Config holds all the game configuration values
type Config struct {
	Title      string
	GameDir    string
	BoardFile  string
	ConfigFile string

	PlayerYOffset  int
	BarrierYOffset int
	AlienYOffset   int

	BaseLevel     int
	BaseLevelStep int
	BaseScore     int
	BaseLives     int

	BaseDifficulty           float64
	BaseDifficultyMultiplier float64

	BasePlayerSize   float64
	BasePlayerSpeed  float64
	BasePlayerHealth float64
	BasePlayerAttack float64

	BaseAliensCount int
	BaseAlienSize   float64
	BaseAlienSpeed  float64
	BaseAlienHealth float64

	BaseProjectileSize   float64
	BaseProjectileSpeed  float64
	BaseProjectileHealth float64

	BaseBarrierSize   float64
	BaseBarrierCount  int
	BaseBarrierHealth float64
	BaseBarrierAttack float64

	BaseShootInterval     float64
	MinShootInterval      float64
	ShootIntervalVariance float64
	BaseShootChance       float64
	CooldownMultiplier    float64
	IntervalRandomFactor  float64
}

func NewConfig(workDir string, baseConfig *Config) (*Config, error) {
	config := &Config{}
	config.GameDir = path.Join(workDir, "spaceinvaders")
	config.ConfigFile = path.Join(config.GameDir, "config.json")
	config.BoardFile = path.Join(config.GameDir, "board.json")

	err := config.Load()
	if err != nil {
		if os.IsNotExist(err) {
			base := baseConfig
			base.GameDir = path.Join(workDir, base.GameDir)
			base.ConfigFile = path.Join(base.GameDir, base.ConfigFile)
			base.BoardFile = path.Join(base.GameDir, base.BoardFile)

			err = base.Save()
			if err != nil {
				return nil, err
			}

			return base, nil
		}

		return nil, err
	}

	return config, nil
}

func (c *Config) Save() error {
	filename := c.ConfigFile
	if err := ensureDir(filename); err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(c)
}

func (c *Config) Load() error {
	filename := c.ConfigFile
	if err := ensureDir(filename); err != nil {
		return err
	}
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(c)
}

func ensureDir(filename string) error {
	dir := filepath.Dir(filename)
	return os.MkdirAll(dir, 0755)
}
