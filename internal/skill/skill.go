package skill

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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

func Copy(srcDir, dstDir string) error {
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dstDir, err)
	}
	return copyDir(srcDir, dstDir)
}

func Remove(dirPath string) error {
	return os.RemoveAll(dirPath)
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

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
