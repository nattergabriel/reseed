package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLibraryConfig_Missing(t *testing.T) {
	cfg, err := LoadLibraryConfig(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Sources) != 0 || len(cfg.Packs) != 0 {
		t.Error("expected empty config for missing file")
	}
}

func TestSaveAndLoadLibraryConfig(t *testing.T) {
	dir := t.TempDir()

	want := &LibraryConfig{
		Sources: map[string]Source{
			"sql-safety": {Source: "github:user/repo/sql-safety", Version: "v1.0"},
		},
		Packs: map[string][]string{
			"backend": {"sql-safety", "logging"},
		},
	}

	if err := SaveLibraryConfig(dir, want); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := LoadLibraryConfig(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got.Sources["sql-safety"] != want.Sources["sql-safety"] {
		t.Errorf("sources: got %+v, want %+v", got.Sources["sql-safety"], want.Sources["sql-safety"])
	}
	if len(got.Packs["backend"]) != 2 {
		t.Errorf("packs: got %v, want %v", got.Packs["backend"], want.Packs["backend"])
	}
}

func TestLoadLibraryConfig_Invalid(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(":::bad yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadLibraryConfig(dir)
	if err == nil {
		t.Fatal("expected error for invalid yaml")
	}
}
