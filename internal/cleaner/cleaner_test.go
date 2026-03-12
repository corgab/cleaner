package cleaner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/corgab/goclean/internal/cleaner"
)

func createDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "node_modules")
	createDir(t, target)

	freed, err := cleaner.Remove(target, false)
	if err != nil {
		t.Fatalf("remove failed: %v", err)
	}
	if freed == 0 {
		t.Error("expected non-zero freed bytes")
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Error("expected directory to be deleted")
	}
}

func TestRemoveDryRun(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "node_modules")
	createDir(t, target)

	freed, err := cleaner.Remove(target, true)
	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}
	if freed == 0 {
		t.Error("expected non-zero freed bytes in dry-run")
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Error("expected directory to still exist after dry-run")
	}
}

func TestRemoveNonexistent(t *testing.T) {
	_, err := cleaner.Remove("/nonexistent/dir", false)
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}
