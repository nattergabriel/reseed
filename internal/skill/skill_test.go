package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func createSkill(t *testing.T, dir, name string) string {
	t.Helper()
	return createSkillWithFrontmatter(t, dir, name, "# "+name)
}

func createSkillWithFrontmatter(t *testing.T, dir, name, content string) string {
	t.Helper()
	skillDir := filepath.Join(dir, name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, MarkerFile), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return skillDir
}

func TestIsSkill(t *testing.T) {
	dir := t.TempDir()

	skillDir := createSkill(t, dir, "my-skill")
	if !IsSkill(skillDir) {
		t.Error("expected true for directory with SKILL.md")
	}

	emptyDir := filepath.Join(dir, "not-a-skill")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if IsSkill(emptyDir) {
		t.Error("expected false for directory without SKILL.md")
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()
	createSkill(t, dir, "b-skill")
	createSkill(t, dir, "a-skill")
	if err := os.MkdirAll(filepath.Join(dir, "not-a-skill"), 0o755); err != nil {
		t.Fatal(err)
	}

	skills, err := List(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 2 || skills[0] != "a-skill" || skills[1] != "b-skill" {
		t.Errorf("got %v, want [a-skill b-skill]", skills)
	}
}

func TestList_NonExistent(t *testing.T) {
	skills, err := List(filepath.Join(t.TempDir(), "nope"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skills != nil {
		t.Errorf("expected nil, got %v", skills)
	}
}

func TestReadDescription(t *testing.T) {
	dir := t.TempDir()

	// Skill with frontmatter description
	withDesc := createSkillWithFrontmatter(t, dir, "with-desc", "---\nname: with-desc\ndescription: A helpful skill\n---\n# With Desc")

	if got := ReadDescription(withDesc); got != "A helpful skill" {
		t.Errorf("got %q, want %q", got, "A helpful skill")
	}

	// Skill without frontmatter
	noFront := createSkill(t, dir, "no-front")

	if got := ReadDescription(noFront); got != "" {
		t.Errorf("got %q, want empty", got)
	}

	// Skill with frontmatter but no description
	noDesc := createSkillWithFrontmatter(t, dir, "no-desc", "---\nname: no-desc\n---\n# No Desc")

	if got := ReadDescription(noDesc); got != "" {
		t.Errorf("got %q, want empty", got)
	}

	// Non-existent directory
	if got := ReadDescription(filepath.Join(dir, "nope")); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestListNested(t *testing.T) {
	dir := t.TempDir()

	createSkill(t, dir, "standalone")
	packDir := filepath.Join(dir, "mypack")
	createSkill(t, packDir, "inpack")

	entries, err := ListNested(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].Name != "standalone" || entries[0].Pack != "" {
		t.Errorf("got %+v, want {Name: standalone, Pack: \"\"}", entries[0])
	}
	if entries[1].Name != "inpack" || entries[1].Pack != "mypack" {
		t.Errorf("got %+v, want {Name: inpack, Pack: mypack}", entries[1])
	}
}

func TestCopy(t *testing.T) {
	src := t.TempDir()
	createSkill(t, src, "my-skill")
	if err := os.WriteFile(filepath.Join(src, "my-skill", "extra.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(t.TempDir(), "my-skill")
	if err := Copy(filepath.Join(src, "my-skill"), dst); err != nil {
		t.Fatalf("copy: %v", err)
	}

	if !IsSkill(dst) {
		t.Error("copied directory should be a skill")
	}
	data, err := os.ReadFile(filepath.Join(dst, "extra.txt"))
	if err != nil || string(data) != "hello" {
		t.Error("extra file not copied correctly")
	}
}
