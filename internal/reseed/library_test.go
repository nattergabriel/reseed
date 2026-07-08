package reseed

import (
	"path/filepath"
	"slices"
	"testing"
)

// testLibrary builds a Library from a temp dir without touching global config.
func testLibrary(t *testing.T) *Library {
	t.Helper()
	dir := t.TempDir()
	createSkill(t, dir, "solo")
	createSkill(t, filepath.Join(dir, "kit"), "commit")
	createSkill(t, filepath.Join(dir, "kit", "frontend"), "react")
	createSkill(t, filepath.Join(dir, "other"), "solo") // duplicate name

	skills, err := FindSkills(dir)
	if err != nil {
		t.Fatal(err)
	}
	return &Library{Path: dir, Skills: skills}
}

func TestResolve(t *testing.T) {
	lib := testLibrary(t)

	tests := []struct {
		arg     string
		want    []string
		wantErr bool
	}{
		{"commit", []string{"commit"}, false},        // unique skill name
		{"kit", []string{"commit", "react"}, false},  // folder, recursive
		{"kit/frontend", []string{"react"}, false},   // nested folder by path
		{"kit/", []string{"commit", "react"}, false}, // trailing slash trimmed
		{"solo", nil, true},                          // ambiguous skill name
		{"nope", nil, true},                          // not found
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			got, err := lib.Resolve(tt.arg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var names []string
			for _, s := range got {
				names = append(names, s.Name)
			}
			if !slices.Equal(names, tt.want) {
				t.Errorf("got %v, want %v", names, tt.want)
			}
		})
	}
}

func TestProjectAddRemoveSync(t *testing.T) {
	lib := testLibrary(t)
	proj := Project{SkillsDir: filepath.Join(t.TempDir(), "skills")}

	commit, err := lib.Resolve("commit")
	if err != nil {
		t.Fatal(err)
	}
	if err := proj.Add(commit[0]); err != nil {
		t.Fatalf("add: %v", err)
	}

	installed, err := proj.Installed()
	if err != nil || len(installed) != 1 || installed[0] != "commit" {
		t.Fatalf("got %v (%v), want [commit]", installed, err)
	}

	updated, err := proj.Sync(lib)
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if len(updated) != 1 || updated[0] != "commit" {
		t.Fatalf("sync updated %v, want [commit]", updated)
	}

	// Syncing an installed skill whose name is ambiguous in the library errors
	// instead of silently picking one.
	createSkill(t, proj.SkillsDir, "solo")
	if _, err := proj.Sync(lib); err == nil {
		t.Error("expected ambiguity error from sync")
	}

	if err := proj.Remove("commit"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := proj.Remove("commit"); err == nil {
		t.Error("expected error removing a skill that is not installed")
	}
}
