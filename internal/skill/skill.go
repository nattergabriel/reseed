package skill

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const MarkerFile = "SKILL.md"

func IsSkill(dirPath string) bool {
	info, err := os.Stat(filepath.Join(dirPath, MarkerFile))
	return err == nil && !info.IsDir()
}

func List(parentDir string) ([]string, error) {
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
	sort.Strings(skills)
	return skills, nil
}

// SkillEntry represents a skill found in the library, potentially inside a pack.
type SkillEntry struct {
	Name        string // leaf directory name, e.g. "commit"
	Pack        string // pack name, empty for standalone skills
	Path        string // full filesystem path to the skill directory
	Description string // from SKILL.md frontmatter, may be empty
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

// ListNested scans parentDir for skills at two levels: standalone skills directly
// in parentDir, and skills inside pack subdirectories. A subdirectory that is not
// itself a skill but contains skill children is treated as a pack.
func ListNested(parentDir string) ([]SkillEntry, error) {
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading directory %s: %w", parentDir, err)
	}

	var skills []SkillEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirPath := filepath.Join(parentDir, e.Name())

		if IsSkill(dirPath) {
			skills = append(skills, SkillEntry{
				Name:        e.Name(),
				Path:        dirPath,
				Description: ReadDescription(dirPath),
			})
			continue
		}

		// Check if this is a pack (non-skill dir containing skills)
		children, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}
		for _, child := range children {
			childPath := filepath.Join(dirPath, child.Name())
			if child.IsDir() && IsSkill(childPath) {
				skills = append(skills, SkillEntry{
					Name:        child.Name(),
					Pack:        e.Name(),
					Path:        childPath,
					Description: ReadDescription(childPath),
				})
			}
		}
	}

	sort.Slice(skills, func(i, j int) bool {
		if skills[i].Pack != skills[j].Pack {
			return skills[i].Pack < skills[j].Pack
		}
		return skills[i].Name < skills[j].Name
	})
	return skills, nil
}

func Copy(srcDir, dstDir string) error {
	if err := os.RemoveAll(dstDir); err != nil {
		return fmt.Errorf("removing old directory %s: %w", dstDir, err)
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dstDir, err)
	}
	return copyDir(srcDir, dstDir)
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())

		if e.IsDir() {
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				return err
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, in.Close()) }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, out.Close()) }()

	_, err = io.Copy(out, in)
	return err
}
