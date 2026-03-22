package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nattergabriel/reseed/internal/library"
	"github.com/nattergabriel/reseed/internal/skill"
)

var DefaultSkillsDir = ".agents/skills"

// SkillsDirOverride is set via the --dir flag on the root command.
var SkillsDirOverride string

func SkillsPath() (string, error) {
	dir := DefaultSkillsDir
	if SkillsDirOverride != "" {
		dir = SkillsDirOverride
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}
	return filepath.Join(cwd, dir), nil
}

func EnsureSkillsDir() (string, error) {
	path, err := SkillsPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", fmt.Errorf("creating %s: %w", path, err)
	}
	return path, nil
}

func ListInstalled() ([]string, error) {
	path, err := SkillsPath()
	if err != nil {
		return nil, err
	}
	return skill.List(path)
}

func AddSkill(lib *library.Library, skillName string) error {
	if !lib.HasSkill(skillName) {
		return fmt.Errorf("skill %q not found in library", skillName)
	}

	projectDir, err := EnsureSkillsDir()
	if err != nil {
		return err
	}

	src := lib.SkillPath(skillName)
	dst := filepath.Join(projectDir, skillName)
	return skill.Copy(src, dst)
}

func SyncSkills(lib *library.Library) ([]string, error) {
	installed, err := ListInstalled()
	if err != nil {
		return nil, err
	}

	projectDir, err := SkillsPath()
	if err != nil {
		return nil, err
	}

	var updated []string
	for _, name := range installed {
		if lib.HasSkill(name) {
			src := lib.SkillPath(name)
			dst := filepath.Join(projectDir, name)
			if err := skill.Copy(src, dst); err != nil {
				return updated, fmt.Errorf("syncing %s: %w", name, err)
			}
			updated = append(updated, name)
		}
	}
	return updated, nil
}
