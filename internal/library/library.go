package library

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nattergabriel/reseed/internal/config"
	"github.com/nattergabriel/reseed/internal/skill"
)

const SkillsSubdir = "skills"

type Library struct {
	Path string
}

func Open() (*Library, error) {
	libPath, err := config.LibraryPath()
	if err != nil {
		return nil, err
	}
	return &Library{Path: libPath}, nil
}

func Init(path string) (*Library, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	skillsDir := filepath.Join(absPath, SkillsSubdir)
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating skills directory: %w", err)
	}

	globalCfg, err := config.LoadGlobal()
	if err != nil {
		globalCfg = &config.GlobalConfig{}
	}
	globalCfg.Library = absPath
	if err := config.SaveGlobal(globalCfg); err != nil {
		return nil, fmt.Errorf("saving global config: %w", err)
	}

	return &Library{Path: absPath}, nil
}

func (l *Library) SkillsDir() string {
	return filepath.Join(l.Path, SkillsSubdir)
}

func (l *Library) ListSkills() ([]string, error) {
	entries, err := l.ListSkillEntries()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}
	return names, nil
}

func (l *Library) ListSkillEntries() ([]skill.SkillEntry, error) {
	return skill.ListAll(l.SkillsDir())
}

// FindSkill locates a skill by name. Skill names must be unique across the
// library; returns an error if ambiguous or not found.
func (l *Library) FindSkill(name string) (*skill.SkillEntry, error) {
	entries, err := l.ListSkillEntries()
	if err != nil {
		return nil, err
	}

	var matches []skill.SkillEntry
	for _, e := range entries {
		if e.Name == name {
			matches = append(matches, e)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("skill %q not found in library", name)
	case 1:
		return &matches[0], nil
	default:
		var locations []string
		for _, m := range matches {
			location := m.Name
			if m.Group != "" {
				location = m.Group + "/" + m.Name
			}
			locations = append(locations, location)
		}
		return nil, fmt.Errorf("skill %q is ambiguous, found at: %v", name, locations)
	}
}

func (l *Library) SkillPath(name string) (string, error) {
	entry, err := l.FindSkill(name)
	if err != nil {
		return "", err
	}
	return entry.Path, nil
}

// Resolve resolves a name to a list of skill names. A name matching a library
// folder (by path relative to skills/) expands to every skill under it, at any
// depth. Otherwise it must match a single skill name.
func (l *Library) Resolve(name string) ([]string, error) {
	entries, err := l.ListSkillEntries()
	if err != nil {
		return nil, err
	}

	name = strings.Trim(name, "/")
	var folderSkills []string
	skillFound := false
	for _, e := range entries {
		if e.Group == name || strings.HasPrefix(e.Group, name+"/") {
			folderSkills = append(folderSkills, e.Name)
		}
		if e.Name == name {
			skillFound = true
		}
	}
	if len(folderSkills) > 0 {
		return folderSkills, nil
	}

	if skillFound {
		return []string{name}, nil
	}

	return nil, fmt.Errorf("%q is not a skill or folder in your library", name)
}
