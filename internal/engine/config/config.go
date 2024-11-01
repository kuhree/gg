package config

import (
	"encoding/json"
	"os"
	"path"
	"strings"
	"unicode"

	"github.com/kuhree/gg/internal/utils"
)

type Config interface {
	Get() *BaseConfig
	Save() error
	Load() error
}

// BaseConfig holds common configuration values and methods
type BaseConfig struct {
	Title      string
	GameDir    string
	BoardFile  string
	ConfigFile string
}

// NewBaseConfig initializes a new BaseConfig
func NewBaseConfig(workDir, gameName string) BaseConfig {
	cleanGameName := strings.ToLower(
		strings.Map(func(r rune) rune {
			switch r {
			case ' ', '-', '|', '\'', ',':
				return -1
			default:
				return unicode.ToLower(r)
			}
		}, gameName),
	)

	return BaseConfig{
		Title:      gameName,
		GameDir:    path.Join(workDir, cleanGameName),
		ConfigFile: path.Join(workDir, cleanGameName, "config.json"),
		BoardFile:  path.Join(workDir, cleanGameName, "board.json"),
	}
}

// Get returns a pointer to the BaseConfig
func (c *BaseConfig) Get() *BaseConfig {
	return c
}

// Save writes the configuration to a file
func (c *BaseConfig) Save() error {
	return SaveConfig(c)
}

// Load reads the configuration from a file
func (c *BaseConfig) Load() error {
	return LoadConfig(c)
}

// SaveConfig is a helper function to save any struct implementing ConfigInterface
func SaveConfig(c Config) error {
	if err := utils.EnsureDir(c.Get().ConfigFile); err != nil {
		return err
	}
	file, err := os.Create(c.Get().ConfigFile)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(c)
}

// LoadConfig is a helper function to load any struct implementing ConfigInterface
func LoadConfig(c Config) error {
	if err := utils.EnsureDir(c.Get().ConfigFile); err != nil {
		return err
	}
	file, err := os.Open(c.Get().ConfigFile)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(c)
}
