// Package config handles loading and saving the Goclean configuration file.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the user's Goclean preferences.
type Config struct {
	Days          int      `yaml:"days"`
	Targets       []string `yaml:"targets"`
	ExcludedPaths []string `yaml:"excluded_paths"`
}

// Default returns a Config with default values and no targets selected.
func Default() Config {
	return Config{
		Days:          30,
		Targets:       []string{},
		ExcludedPaths: []string{},
	}
}

// Exists reports whether a config file exists at the given path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Load reads and parses the YAML config file at the given path.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Save writes the config as YAML to the given path.
func Save(cfg Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
