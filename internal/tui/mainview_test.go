package tui_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/corgab/goclean/internal/scanner"
	"github.com/corgab/goclean/internal/tui"
)

func sampleResults() []scanner.Result {
	return []scanner.Result{
		{
			Path:          "/tmp/project1/node_modules",
			ProjectDir:    "/tmp/project1",
			DependencyDir: "node_modules",
			TargetName:    "Node.js",
			Size:          104857600,
			ModTime:       time.Now().Add(-60 * 24 * time.Hour),
			Stale:         true,
		},
		{
			Path:          "/tmp/project2/vendor",
			ProjectDir:    "/tmp/project2",
			DependencyDir: "vendor",
			TargetName:    "PHP (Composer)",
			Size:          52428800,
			ModTime:       time.Now().Add(-90 * 24 * time.Hour),
			Stale:         true,
		},
	}
}

func TestMainModelInitialState(t *testing.T) {
	m := tui.NewMainModel(sampleResults(), false, "")
	if m.DeletedCount() != 0 {
		t.Error("expected 0 deleted initially")
	}
	if m.FreedBytes() != 0 {
		t.Error("expected 0 freed initially")
	}
	if len(m.SelectedPaths()) != 0 {
		t.Error("expected no selections initially")
	}
}

func TestMainModelToggleSelection(t *testing.T) {
	m := tui.NewMainModel(sampleResults(), false, "")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(tui.MainModel)

	selected := m.SelectedPaths()
	if len(selected) != 1 {
		t.Fatalf("expected 1 selected, got %d", len(selected))
	}
	if selected[0] != "/tmp/project1/node_modules" {
		t.Errorf("unexpected path: %s", selected[0])
	}
}

func TestMainModelSelectAll(t *testing.T) {
	m := tui.NewMainModel(sampleResults(), false, "")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(tui.MainModel)

	if len(m.SelectedPaths()) != 2 {
		t.Errorf("expected 2 selected, got %d", len(m.SelectedPaths()))
	}
}

func TestMainModelDeselectAll(t *testing.T) {
	m := tui.NewMainModel(sampleResults(), false, "")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(tui.MainModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(tui.MainModel)

	if len(m.SelectedPaths()) != 0 {
		t.Errorf("expected 0 selected after deselect all, got %d", len(m.SelectedPaths()))
	}
}

func TestMainModelEmptyResults(t *testing.T) {
	m := tui.NewMainModel([]scanner.Result{}, false, "")
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view even with no results")
	}
}

func TestMainModelViewContainsStats(t *testing.T) {
	m := tui.NewMainModel(sampleResults(), false, "Goclean has freed 5.0 GB in 10 operations")
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}
