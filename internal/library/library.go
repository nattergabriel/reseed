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
	Path   string
	Config *config.LibraryConfig
}

func Open() (*Library, error) {
	libPath, err := config.LibraryPath()
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadLibraryConfig(libPath)
	if err != nil {
		return nil, err
	}

	return &Library{Path: libPath, Config: cfg}, nil
}

func Init(path string) (*Library, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	skillsDir := filepath.Join(absPath, SkillsSubdir)
	if err := createDirIfNotExists(skillsDir); err != nil {
		return nil, fmt.Errorf("creating skills directory: %w", err)
	}

	cfg, err := config.LoadLibraryConfig(absPath)
	if err != nil {
		return nil, err
	}

	if err := config.SaveLibraryConfig(absPath, cfg); err != nil {
		return nil, err
	}

	if err := config.SaveGlobal(&config.GlobalConfig{Library: absPath}); err != nil {
		return nil, err
	}

	return &Library{Path: absPath, Config: cfg}, nil
}

func (l *Library) SkillsDir() string {
	return filepath.Join(l.Path, SkillsSubdir)
}

func (l *Library) SkillPath(name string) string {
	return filepath.Join(l.Path, SkillsSubdir, name)
}

func (l *Library) ListSkills() ([]string, error) {
	return skill.List(l.SkillsDir())
}

func (l *Library) HasSkill(name string) bool {
	return skill.IsSkill(l.SkillPath(name))
}

func (l *Library) IsExternal(name string) bool {
	_, ok := l.Config.Sources[name]
	return ok
}

func (l *Library) ResolvePack(name string) ([]string, error) {
	skills, ok := l.Config.Packs[name]
	if !ok {
		return nil, fmt.Errorf("pack %q not found", name)
	}
	return skills, nil
}

func (l *Library) ResolveSkillOrPack(name string) ([]string, error) {
	if skills, ok := l.Config.Packs[name]; ok {
		return skills, nil
	}
	if l.HasSkill(name) {
		return []string{name}, nil
	}
	if _, ok := l.Config.Sources[name]; ok {
		return []string{name}, nil
	}
	return nil, fmt.Errorf("%q is not a skill or pack in your library", name)
}

func (l *Library) SaveConfig() error {
	return config.SaveLibraryConfig(l.Path, l.Config)
}

func createDirIfNotExists(path string) error {
	return os.MkdirAll(path, 0o755)
}
