package targets_test

import (
	"testing"

	"github.com/corgab/goclean/pkg/targets"
)

func TestAllTargetsHaveRequiredFields(t *testing.T) {
	all := targets.All()
	if len(all) == 0 {
		t.Fatal("expected at least one target, got none")
	}
	for _, tgt := range all {
		if tgt.Name == "" {
			t.Error("target has empty Name")
		}
		if tgt.DependencyDir == "" {
			t.Errorf("target %q has empty DependencyDir", tgt.Name)
		}
		if tgt.ConfigFile == "" {
			t.Errorf("target %q has empty ConfigFile", tgt.Name)
		}
	}
}

func TestGetByDirName(t *testing.T) {
	tgt, ok := targets.GetByDirName("node_modules")
	if !ok {
		t.Fatal("expected to find target for node_modules")
	}
	if tgt.Name != "Node.js" {
		t.Errorf("expected name Node.js, got %q", tgt.Name)
	}
	if tgt.ConfigFile != "package.json" {
		t.Errorf("expected config file package.json, got %q", tgt.ConfigFile)
	}

	_, ok = targets.GetByDirName("nonexistent")
	if ok {
		t.Error("expected not to find target for nonexistent dir")
	}
}

func TestGetAllByDirName(t *testing.T) {
	all := targets.GetAllByDirName("vendor")
	if len(all) < 2 {
		t.Fatalf("expected at least 2 targets for vendor, got %d", len(all))
	}
	names := make(map[string]bool)
	for _, tgt := range all {
		names[tgt.Name] = true
	}
	if !names["PHP (Composer)"] {
		t.Error("expected PHP (Composer) in vendor targets")
	}
	if !names["Go"] {
		t.Error("expected Go in vendor targets")
	}

	all = targets.GetAllByDirName("target")
	if len(all) < 2 {
		t.Fatalf("expected at least 2 targets for target, got %d", len(all))
	}

	all = targets.GetAllByDirName("nonexistent")
	if len(all) != 0 {
		t.Errorf("expected 0 targets for nonexistent, got %d", len(all))
	}
}

func TestDirNames(t *testing.T) {
	names := targets.DirNames()
	if len(names) == 0 {
		t.Fatal("expected at least one dir name")
	}
	found := false
	for _, n := range names {
		if n == "node_modules" {
			found = true
		}
	}
	if !found {
		t.Error("expected node_modules in dir names")
	}
}

func TestForDirs(t *testing.T) {
	filtered := targets.ForDirs([]string{"node_modules", "vendor"})
	if len(filtered) == 0 {
		t.Fatal("expected at least one target for node_modules+vendor")
	}
	hasNode := false
	hasVendor := false
	for _, tgt := range filtered {
		if tgt.DependencyDir == "node_modules" {
			hasNode = true
		}
		if tgt.DependencyDir == "vendor" {
			hasVendor = true
		}
	}
	if !hasNode {
		t.Error("expected node_modules target in filtered results")
	}
	if !hasVendor {
		t.Error("expected vendor target in filtered results")
	}

	empty := targets.ForDirs([]string{})
	if len(empty) != 0 {
		t.Errorf("expected 0 targets for empty input, got %d", len(empty))
	}
}
