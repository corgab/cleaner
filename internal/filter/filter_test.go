package filter_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/corgab/cleaner/internal/filter"
)

func TestIsStale_OldFile(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "package.json")
	if err := os.WriteFile(configFile, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-60 * 24 * time.Hour)
	if err := os.Chtimes(configFile, old, old); err != nil {
		t.Fatal(err)
	}

	stale, modTime, err := filter.IsStale(configFile, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stale {
		t.Error("expected file to be stale")
	}
	if modTime.After(time.Now().Add(-59 * 24 * time.Hour)) {
		t.Errorf("expected mod time ~60 days ago, got %v", modTime)
	}
}

func TestIsStale_RecentFile(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "package.json")
	if err := os.WriteFile(configFile, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	stale, _, err := filter.IsStale(configFile, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stale {
		t.Error("expected file NOT to be stale")
	}
}

func TestIsStale_MissingFile(t *testing.T) {
	_, _, err := filter.IsStale("/nonexistent/package.json", 30)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestIsStale_BoundaryExactlyAtThreshold(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "package.json")
	if err := os.WriteFile(configFile, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	boundary := time.Now().Add(-30 * 24 * time.Hour)
	if err := os.Chtimes(configFile, boundary, boundary); err != nil {
		t.Fatal(err)
	}

	stale, _, err := filter.IsStale(configFile, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stale {
		t.Error("expected file at exactly threshold to be stale")
	}
}
