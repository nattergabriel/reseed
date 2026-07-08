package reseed

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const MarkerFile = "SKILL.md"

// Skill is a directory containing a SKILL.md file.
type Skill struct {
	Name  string // leaf directory name, e.g. "commit"
	Group string // containing folder relative to the skills root, "" for top-level skills
	Path  string // absolute path to the skill directory
}

// location is the skill's path relative to the skills root, e.g. "kit/commit".
func (s Skill) location() string {
	if s.Group == "" {
		return s.Name
	}
	return s.Group + "/" + s.Name
}

func IsSkill(dirPath string) bool {
	info, err := os.Stat(filepath.Join(dirPath, MarkerFile))
	return err == nil && !info.IsDir()
}

// listSkills returns the names of skills directly inside parentDir (flat, no
// recursion). This is how a project's skills directory is read.
func listSkills(parentDir string) ([]string, error) {
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading directory %s: %w", parentDir, err)
	}

	var skills []string
	for _, e := range entries {
		if e.IsDir() && IsSkill(filepath.Join(parentDir, e.Name())) {
			skills = append(skills, e.Name())
		}
	}
	return skills, nil
}

// FindSkills walks root recursively and returns every skill found at any
// depth. Folders above a skill are just organization; the walk does not
// descend into skill directories themselves.
func FindSkills(root string) ([]Skill, error) {
	var skills []Skill
	var walk func(dir, group string) error
	walk = func(dir, group string) error {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("reading directory %s: %w", dir, err)
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			dirPath := filepath.Join(dir, e.Name())
			if IsSkill(dirPath) {
				skills = append(skills, Skill{
					Name:  e.Name(),
					Group: group,
					Path:  dirPath,
				})
				continue
			}
			childGroup := e.Name()
			if group != "" {
				childGroup = group + "/" + e.Name()
			}
			if err := walk(dirPath, childGroup); err != nil {
				return err
			}
		}
		return nil
	}

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, nil
	}
	if err := walk(root, ""); err != nil {
		return nil, err
	}

	sort.Slice(skills, func(i, j int) bool {
		if skills[i].Group != skills[j].Group {
			return skills[i].Group < skills[j].Group
		}
		return skills[i].Name < skills[j].Name
	})
	return skills, nil
}

// ReadDescription extracts the description field from a SKILL.md frontmatter.
// Returns an empty string if the file has no frontmatter or no description.
func ReadDescription(skillDir string) string {
	f, err := os.Open(filepath.Join(skillDir, MarkerFile))
	if err != nil {
		return ""
	}
	defer f.Close() //nolint:errcheck // best-effort read

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return ""
	}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			break
		}
		if strings.HasPrefix(line, "description:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
	}
	return ""
}

// CopySkill replaces dstDir with a full copy of srcDir.
func CopySkill(srcDir, dstDir string) error {
	if err := os.RemoveAll(dstDir); err != nil {
		return fmt.Errorf("removing old directory %s: %w", dstDir, err)
	}
	return os.CopyFS(dstDir, os.DirFS(srcDir))
}
