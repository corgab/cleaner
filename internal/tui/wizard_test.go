package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/corgab/cleaner/internal/tui"
)

func TestWizardInitialState(t *testing.T) {
	m := tui.NewWizardModel()
	if m.Done() {
		t.Error("expected wizard not done initially")
	}
	if m.Quitting() {
		t.Error("expected wizard not quitting initially")
	}
	if len(m.WizardResult()) != 0 {
		t.Error("expected no selections initially")
	}
}

func TestWizardToggleAndConfirm(t *testing.T) {
	m := tui.NewWizardModel()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(tui.WizardModel)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(tui.WizardModel)

	if !m.Done() {
		t.Error("expected wizard to be done after enter")
	}
	result := m.WizardResult()
	if len(result) != 1 {
		t.Fatalf("expected 1 selection, got %d", len(result))
	}
}

func TestWizardSelectAll(t *testing.T) {
	m := tui.NewWizardModel()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(tui.WizardModel)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(tui.WizardModel)

	if !m.Done() {
		t.Error("expected wizard to be done")
	}
	if len(m.WizardResult()) == 0 {
		t.Error("expected all targets selected")
	}
}

func TestWizardQuit(t *testing.T) {
	m := tui.NewWizardModel()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(tui.WizardModel)

	if !m.Quitting() {
		t.Error("expected wizard to be quitting after q")
	}
}

func TestWizardEscQuit(t *testing.T) {
	m := tui.NewWizardModel()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(tui.WizardModel)

	if !m.Quitting() {
		t.Error("expected wizard to be quitting after esc")
	}
}
