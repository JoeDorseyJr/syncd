package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Taps    []string `yaml:"taps"`
	Brews   []string `yaml:"brews"`
	Casks   []string `yaml:"casks"`
	Pin     []string `yaml:"pin"`
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
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config in %s: %w", path, err)
	}

	return &cfg, nil
}
