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

// SkillEntry represents a skill found in the library.
type SkillEntry struct {
	Name  string // leaf directory name, e.g. "commit"
	Group string // relative path of the containing folder, empty for top-level skills
	Path  string // full filesystem path to the skill directory
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

// ListAll walks parentDir recursively and returns every skill found at any
// depth. Folders above a skill are just organization; the walk does not
// descend into skill directories themselves.
func ListAll(parentDir string) ([]SkillEntry, error) {
	var skills []SkillEntry
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
				skills = append(skills, SkillEntry{
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

	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		return nil, nil
	}
	if err := walk(parentDir, ""); err != nil {
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
