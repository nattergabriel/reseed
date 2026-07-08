package reseed

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const SkillsSubdir = "skills"

// Library is the source side: the user's central skill collection, walked
// once at open time.
type Library struct {
	Path   string  // library root
	Skills []Skill // every skill in the library
}

// OpenLibrary opens the library recorded in the global config.
func OpenLibrary() (*Library, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	if cfg.Library == "" {
		return nil, fmt.Errorf("library not initialized: run 'reseed init <path>' first")
	}
	return openLibraryAt(cfg.Library)
}

// InitLibrary creates (or recognizes) a library at path and records it in the
// global config.
func InitLibrary(path string) (*Library, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(absPath, SkillsSubdir), 0o755); err != nil {
		return nil, fmt.Errorf("creating skills directory: %w", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	cfg.Library = absPath
	if err := SaveConfig(cfg); err != nil {
		return nil, fmt.Errorf("saving global config: %w", err)
	}

	return openLibraryAt(absPath)
}

func openLibraryAt(path string) (*Library, error) {
	skillsDir := filepath.Join(path, SkillsSubdir)
	if _, err := os.Stat(skillsDir); err != nil {
		return nil, fmt.Errorf("library at %s has no skills directory: run 'reseed init <path>' to fix", path)
	}
	skills, err := FindSkills(skillsDir)
	if err != nil {
		return nil, err
	}
	return &Library{Path: path, Skills: skills}, nil
}

func (l *Library) SkillsDir() string {
	return filepath.Join(l.Path, SkillsSubdir)
}

// Resolve resolves a name to skills. A name matching a library folder (by
// path relative to the skills root) expands to every skill under it, at any
// depth. Otherwise it must match a unique skill name.
func (l *Library) Resolve(name string) ([]Skill, error) {
	name = strings.Trim(name, "/")
	var folder, named []Skill
	for _, s := range l.Skills {
		if s.Group == name || strings.HasPrefix(s.Group, name+"/") {
			folder = append(folder, s)
		}
		if s.Name == name {
			named = append(named, s)
		}
	}
	if len(folder) > 0 {
		return folder, nil
	}

	switch len(named) {
	case 0:
		return nil, fmt.Errorf("%q is not a skill or folder in your library", name)
	case 1:
		return named, nil
	default:
		return nil, ambiguousSkillError(name, named)
	}
}

func ambiguousSkillError(name string, matches []Skill) error {
	var locations []string
	for _, m := range matches {
		locations = append(locations, m.location())
	}
	return fmt.Errorf("skill %q is ambiguous, found at: %v", name, locations)
}
