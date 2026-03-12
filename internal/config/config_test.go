package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/corgab/goclean/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.Default()
	if cfg.Days != 30 {
		t.Errorf("expected default days 30, got %d", cfg.Days)
	}
	if len(cfg.Targets) != 0 {
		t.Errorf("expected no default targets, got %v", cfg.Targets)
	}
	if len(cfg.ExcludedPaths) != 0 {
		t.Errorf("expected no excluded paths, got %v", cfg.ExcludedPaths)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := config.Config{
		Days:          60,
		Targets:       []string{"node_modules", "vendor"},
		ExcludedPaths: []string{"/tmp/keep"},
	}

	if err := config.Save(cfg, path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Days != 60 {
		t.Errorf("expected days 60, got %d", loaded.Days)
	}
	if len(loaded.Targets) != 2 || loaded.Targets[0] != "node_modules" {
		t.Errorf("unexpected targets: %v", loaded.Targets)
	}
	if len(loaded.ExcludedPaths) != 1 || loaded.ExcludedPaths[0] != "/tmp/keep" {
		t.Errorf("unexpected excluded paths: %v", loaded.ExcludedPaths)
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error loading nonexistent config")
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if config.Exists(path) {
		t.Error("expected Exists to return false for missing file")
	}

	if err := os.WriteFile(path, []byte("days: 30\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if !config.Exists(path) {
		t.Error("expected Exists to return true for existing file")
	}
}
