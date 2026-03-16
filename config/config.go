package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const defaultBaseURL = "http://localhost:3030"

type Config struct {
	BaseURL string `json:"base_url"`
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".jeeves"), nil
}

func EnsureDir() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0700)
}

func Load() (*Config, error) {
	dir, err := Dir()
	if err != nil {
		return &Config{BaseURL: defaultBaseURL}, nil
	}

	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{BaseURL: defaultBaseURL}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	return &cfg, nil
}
