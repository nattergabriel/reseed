package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type GlobalConfig struct {
	Library string `yaml:"library"`
}

func GlobalConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving config dir: %w", err)
	}
	return filepath.Join(dir, "reseed"), nil
}

func LoadGlobal() (*GlobalConfig, error) {
	path, err := GlobalConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("library not initialized: run 'reseed init <path>' first")
		}
		return nil, fmt.Errorf("reading global config: %w", err)
	}

	var cfg GlobalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing global config: %w", err)
	}

	if cfg.Library == "" {
		return nil, fmt.Errorf("library not initialized: run 'reseed init <path>' first")
	}

	return &cfg, nil
}

func SaveGlobal(cfg *GlobalConfig) error {
	path, err := GlobalConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling global config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing global config: %w", err)
	}

	return nil
}

func LibraryPath() (string, error) {
	cfg, err := LoadGlobal()
	if err != nil {
		return "", err
	}
	return cfg.Library, nil
}
