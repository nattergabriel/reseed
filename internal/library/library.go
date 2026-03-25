package library

import (
	"fmt"
	"os"
	"path/filepath"

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
		return nil, err
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
	return skill.ListNested(l.SkillsDir())
}

// FindSkill locates a skill by name, preferring standalone skills over
// pack members. Returns an error if ambiguous or not found.
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
		// Check if one is standalone (preferred)
		for _, m := range matches {
			if m.Pack == "" {
				return &m, nil
			}
		}
		var packs []string
		for _, m := range matches {
			packs = append(packs, m.Pack)
		}
		return nil, fmt.Errorf("skill %q is ambiguous, found in packs: %v", name, packs)
	}
}

func (l *Library) SkillPath(name string) (string, error) {
	entry, err := l.FindSkill(name)
	if err != nil {
		return "", err
	}
	return entry.Path, nil
}

// ResolveSkillOrPack resolves a name to a list of skill names.
// If the name matches a pack directory, returns all skills in that pack.
// Otherwise, returns the single skill name.
func (l *Library) ResolveSkillOrPack(name string) ([]string, error) {
	entries, err := l.ListSkillEntries()
	if err != nil {
		return nil, err
	}

	// Check if it's a pack
	var packSkills []string
	found := false
	for _, e := range entries {
		if e.Pack == name {
			packSkills = append(packSkills, e.Name)
			found = true
		} else if e.Name == name {
			found = true
		}
	}
	if len(packSkills) > 0 {
		return packSkills, nil
	}

	if found {
		return []string{name}, nil
	}

	return nil, fmt.Errorf("%q is not a skill or pack in your library", name)
}
