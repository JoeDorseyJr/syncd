package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Taps    []string `yaml:"taps"`
	Brews   []string `yaml:"brews"`
	Casks   []string `yaml:"casks"`
	Cleanup Cleanup  `yaml:"cleanup"`
}

type Cleanup struct {
	RemoveUnlisted bool `yaml:"remove_unlisted"`
	ClearCache     bool `yaml:"clear_cache"`
	Autoremove     bool `yaml:"autoremove"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("determining home directory: %w", err)
		}
		path = filepath.Join(home, ".config", "syncd", "config.yaml")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML in %s: %w", path, err)
	}

	return &cfg, nil
}
