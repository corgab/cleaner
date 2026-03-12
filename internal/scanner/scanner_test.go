package scanner_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/corgab/cleaner/internal/scanner"
)

func createFakeProject(t *testing.T, base, projectName, depDir, configFile string, modTime time.Time) {
	t.Helper()
	projectDir := filepath.Join(base, projectName)
	depPath := filepath.Join(projectDir, depDir)
	if err := os.MkdirAll(depPath, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(depPath, "dummy.txt"), make([]byte, 1024), 0644); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(projectDir, configFile)
	if err := os.WriteFile(cfgPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(cfgPath, modTime, modTime); err != nil {
		t.Fatal(err)
	}
}

func TestScanFindsStaleProject(t *testing.T) {
	base := t.TempDir()
	oldTime := time.Now().Add(-60 * 24 * time.Hour)
	createFakeProject(t, base, "myproject", "node_modules", "package.json", oldTime)

	results, err := scanner.Scan(scanner.Options{
		Root:          base,
		Days:          30,
		Targets:       []string{"node_modules"},
		ExcludedPaths: []string{},
	})
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.DependencyDir != "node_modules" {
		t.Errorf("expected dep dir node_modules, got %q", r.DependencyDir)
	}
	if r.TargetName != "Node.js" {
		t.Errorf("expected target name Node.js, got %q", r.TargetName)
	}
	if r.Size == 0 {
		t.Error("expected non-zero size")
	}
	if !r.Stale {
		t.Error("expected project to be stale")
	}
	if r.ProjectDir != filepath.Join(base, "myproject") {
		t.Errorf("expected project dir %s, got %s", filepath.Join(base, "myproject"), r.ProjectDir)
	}
}

func TestScanSkipsRecentProject(t *testing.T) {
	base := t.TempDir()
	recentTime := time.Now()
	createFakeProject(t, base, "myproject", "node_modules", "package.json", recentTime)

	results, err := scanner.Scan(scanner.Options{
		Root:          base,
		Days:          30,
		Targets:       []string{"node_modules"},
		ExcludedPaths: []string{},
	})
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for recent project, got %d", len(results))
	}
}

func TestScanSkipsDotDirectories(t *testing.T) {
	base := t.TempDir()
	hiddenDir := filepath.Join(base, ".hidden", "project")
	depPath := filepath.Join(hiddenDir, "node_modules")
	if err := os.MkdirAll(depPath, 0755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(hiddenDir, "package.json")
	old := time.Now().Add(-60 * 24 * time.Hour)
	if err := os.WriteFile(cfgPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(cfgPath, old, old); err != nil {
		t.Fatal(err)
	}

	results, err := scanner.Scan(scanner.Options{
		Root:          base,
		Days:          30,
		Targets:       []string{"node_modules"},
		ExcludedPaths: []string{},
	})
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results (hidden dir skipped), got %d", len(results))
	}
}

func TestScanSkipsExcludedPaths(t *testing.T) {
	base := t.TempDir()
	oldTime := time.Now().Add(-60 * 24 * time.Hour)
	createFakeProject(t, base, "myproject", "node_modules", "package.json", oldTime)

	results, err := scanner.Scan(scanner.Options{
		Root:          base,
		Days:          30,
		Targets:       []string{"node_modules"},
		ExcludedPaths: []string{filepath.Join(base, "myproject")},
	})
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results (excluded path), got %d", len(results))
	}
}

func TestScanMissingConfigFileSkipped(t *testing.T) {
	base := t.TempDir()
	// Create a dep dir without any config file in the parent.
	// This should NOT be reported (avoids false positives like
	// Laravel's resources/views/vendor).
	depPath := filepath.Join(base, "orphan", "node_modules")
	if err := os.MkdirAll(depPath, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(depPath, "f.txt"), make([]byte, 512), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := scanner.Scan(scanner.Options{
		Root:          base,
		Days:          30,
		Targets:       []string{"node_modules"},
		ExcludedPaths: []string{},
	})
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for dep dir without config file, got %d", len(results))
	}
}

func TestScanDisambiguatesSharedDirNames(t *testing.T) {
	base := t.TempDir()
	oldTime := time.Now().Add(-60 * 24 * time.Hour)

	createFakeProject(t, base, "goproject", "vendor", "go.mod", oldTime)
	createFakeProject(t, base, "phpproject", "vendor", "composer.json", oldTime)

	results, err := scanner.Scan(scanner.Options{
		Root:          base,
		Days:          30,
		Targets:       []string{"vendor"},
		ExcludedPaths: []string{},
	})
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	names := make(map[string]bool)
	for _, r := range results {
		names[r.TargetName] = true
	}
	if !names["Go"] {
		t.Error("expected Go target to be identified")
	}
	if !names["PHP (Composer)"] {
		t.Error("expected PHP (Composer) target to be identified")
	}
}

func TestScanMultipleTargetTypes(t *testing.T) {
	base := t.TempDir()
	oldTime := time.Now().Add(-60 * 24 * time.Hour)

	createFakeProject(t, base, "nodeproject", "node_modules", "package.json", oldTime)
	createFakeProject(t, base, "goproject", "vendor", "go.mod", oldTime)

	results, err := scanner.Scan(scanner.Options{
		Root:          base,
		Days:          30,
		Targets:       []string{"node_modules", "vendor"},
		ExcludedPaths: []string{},
	})
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}
