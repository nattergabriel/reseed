package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Source struct {
	Source  string `yaml:"source"`
	Version string `yaml:"version"`
}

type LibraryConfig struct {
	Sources map[string]Source   `yaml:"sources,omitempty"`
	Packs   map[string][]string `yaml:"packs,omitempty"`
}

func LoadLibraryConfig(libraryPath string) (*LibraryConfig, error) {
	path := filepath.Join(libraryPath, "reseed.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &LibraryConfig{
				Sources: make(map[string]Source),
				Packs:   make(map[string][]string),
			}, nil
		}
		return nil, fmt.Errorf("reading library config: %w", err)
	}

	var cfg LibraryConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing library config: %w", err)
	}

	if cfg.Sources == nil {
		cfg.Sources = make(map[string]Source)
	}
	if cfg.Packs == nil {
		cfg.Packs = make(map[string][]string)
	}

	return &cfg, nil
}

func SaveLibraryConfig(libraryPath string, cfg *LibraryConfig) error {
	path := filepath.Join(libraryPath, "reseed.yaml")

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling library config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing library config: %w", err)
	}

	return nil
}
