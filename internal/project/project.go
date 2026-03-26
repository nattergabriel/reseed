package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nattergabriel/reseed/internal/config"
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
	} else if cfg, err := config.LoadGlobal(); err == nil && cfg.Dir != "" {
		dir = cfg.Dir
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

func InstalledSet() (map[string]bool, error) {
	names, err := ListInstalled()
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	return set, nil
}

func AddSkill(lib *library.Library, skillName string) error {
	srcPath, err := lib.SkillPath(skillName)
	if err != nil {
		return err
	}

	projectDir, err := EnsureSkillsDir()
	if err != nil {
		return err
	}

	dst := filepath.Join(projectDir, skillName)
	return skill.Copy(srcPath, dst)
}

func RemoveSkill(skillName string) error {
	projectDir, err := SkillsPath()
	if err != nil {
		return err
	}
	dst := filepath.Join(projectDir, skillName)
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("removing %s: %w", skillName, err)
	}
	return nil
}

func SyncSkills(lib *library.Library) ([]string, error) {
	projectDir, err := SkillsPath()
	if err != nil {
		return nil, err
	}

	installed, err := skill.List(projectDir)
	if err != nil {
		return nil, err
	}

	entries, err := lib.ListSkillEntries()
	if err != nil {
		return nil, err
	}
	pathByName := make(map[string]string, len(entries))
	for _, e := range entries {
		if _, exists := pathByName[e.Name]; !exists {
			pathByName[e.Name] = e.Path
		}
	}

	var updated []string
	for _, name := range installed {
		srcPath, ok := pathByName[name]
		if !ok {
			continue
		}
		dst := filepath.Join(projectDir, name)
		if err := skill.Copy(srcPath, dst); err != nil {
			return updated, fmt.Errorf("syncing %s: %w", name, err)
		}
		updated = append(updated, name)
	}
	return updated, nil
}
