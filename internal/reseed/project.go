package reseed

import (
	"fmt"
	"os"
	"path/filepath"
)

const DefaultSkillsDir = ".agents/skills"

// Project is the destination side: the current project's skills directory.
type Project struct {
	SkillsDir string // absolute
}

// OpenProject resolves the project's skills directory once: the --dir
// override, then the global config, then the default.
func OpenProject(override string) (Project, error) {
	dir := override
	if dir == "" {
		cfg, err := LoadConfig()
		if err != nil {
			return Project{}, err
		}
		dir = cfg.Dir
	}
	if dir == "" {
		dir = DefaultSkillsDir
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Project{}, fmt.Errorf("resolving path: %w", err)
	}
	return Project{SkillsDir: abs}, nil
}

// Installed lists the names of skills in the project.
func (p Project) Installed() ([]string, error) {
	return listSkills(p.SkillsDir)
}

// Add copies a library skill into the project.
func (p Project) Add(s Skill) error {
	if err := os.MkdirAll(p.SkillsDir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", p.SkillsDir, err)
	}
	return CopySkill(s.Path, filepath.Join(p.SkillsDir, s.Name))
}

// Remove deletes a skill directory from the project. It works even when the
// directory is no longer a valid skill (e.g. its SKILL.md was deleted).
func (p Project) Remove(name string) error {
	path := filepath.Join(p.SkillsDir, name)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill %q not installed", name)
		}
		return err
	}
	return os.RemoveAll(path)
}

// Sync re-copies every installed skill that exists in the library. Skills not
// in the library are left untouched; a name matching several library skills
// is an error.
func (p Project) Sync(lib *Library) ([]string, error) {
	installed, err := p.Installed()
	if err != nil {
		return nil, err
	}

	var updated []string
	for _, name := range installed {
		matches, err := lib.matchByName(name)
		if err != nil {
			return updated, err
		}
		if len(matches) == 0 {
			continue
		}
		if err := CopySkill(matches[0].Path, filepath.Join(p.SkillsDir, name)); err != nil {
			return updated, fmt.Errorf("syncing %s: %w", name, err)
		}
		updated = append(updated, name)
	}
	return updated, nil
}
