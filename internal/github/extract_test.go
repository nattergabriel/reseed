package github

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

// buildTarball builds a tar.gz stream shaped like a GitHub repo tarball:
// every entry lives under a "{owner}-{repo}-{sha}/" root directory.
func buildTarball(t *testing.T, files map[string]string) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{
			Name:     "owner-repo-abc123/" + name,
			Mode:     0o644,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf
}

var testRepoFiles = map[string]string{
	"README.md":                         "readme",
	"skills/commit/SKILL.md":            "commit skill",
	"skills/commit/extra.txt":           "extra",
	"skills/nested/deep/react/SKILL.md": "react skill",
	"other/tool/SKILL.md":               "tool skill",
}

func TestExtractSkills(t *testing.T) {
	tests := []struct {
		name    string
		filter  string
		want    []string
		wantErr bool
	}{
		{"whole repo", "", []string{"commit", "react", "tool"}, false},
		{"folder filter", "skills", []string{"commit", "react"}, false},
		{"single skill", "skills/commit", []string{"commit"}, false},
		{"no match", "nope", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destDir := t.TempDir()
			got, err := extractSkills(buildTarball(t, testRepoFiles), destDir, tt.filter)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}

			// Skills are flattened into destDir/<name>/ with their contents.
			for _, name := range tt.want {
				if _, err := os.Stat(filepath.Join(destDir, name, "SKILL.md")); err != nil {
					t.Errorf("skill %s not extracted: %v", name, err)
				}
			}
			if contains(tt.want, "commit") {
				data, err := os.ReadFile(filepath.Join(destDir, "commit", "extra.txt"))
				if err != nil || string(data) != "extra" {
					t.Errorf("supporting file not extracted: %v", err)
				}
			}
		})
	}
}

func TestExtractSkills_NoSkills(t *testing.T) {
	tarball := buildTarball(t, map[string]string{"README.md": "readme"})
	if _, err := extractSkills(tarball, t.TempDir(), ""); err == nil {
		t.Fatal("expected error for a repo containing no skills")
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
