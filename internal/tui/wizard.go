// Package tui implements the Bubbletea-based terminal UI for Goclean.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/corgab/goclean/pkg/targets"
)

var (
	wizardTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
	wizardItemStyle     = lipgloss.NewStyle().PaddingLeft(2)
	wizardSelectedStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("120"))
	wizardCursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

// WizardModel is the Bubbletea model for the first-run target selection wizard.
type WizardModel struct {
	items    []wizardItem
	cursor   int
	done     bool
	quitting bool
}

type wizardItem struct {
	target   targets.Target
	selected bool
}

// WizardResult returns the selected dependency dir names after the wizard completes.
func (m WizardModel) WizardResult() []string {
	var selected []string
	for _, item := range m.items {
		if item.selected {
			selected = append(selected, item.target.DependencyDir)
		}
	}
	return selected
}

// Done reports whether the wizard has completed.
func (m WizardModel) Done() bool { return m.done }

// Quitting reports whether the user quit the wizard.
func (m WizardModel) Quitting() bool { return m.quitting }

// NewWizardModel creates a new wizard model with all available targets.
func NewWizardModel() WizardModel {
	all := targets.All()
	items := make([]wizardItem, len(all))
	for i, t := range all {
		items[i] = wizardItem{target: t}
	}
	return WizardModel{items: items}
}

func (m WizardModel) Init() tea.Cmd { return nil }

func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case tea.KeySpace:
			m.items[m.cursor].selected = !m.items[m.cursor].selected
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "q":
				m.quitting = true
				return m, tea.Quit
			case "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "j":
				if m.cursor < len(m.items)-1 {
					m.cursor++
				}
			case "a":
				allSelected := true
				for _, item := range m.items {
					if !item.selected {
						allSelected = false
						break
					}
				}
				for i := range m.items {
					m.items[i].selected = !allSelected
				}
			}
		}
	}
	return m, nil
}

func (m WizardModel) View() string {
	var b strings.Builder

	b.WriteString(wizardTitleStyle.Render("Goclean - Setup Wizard"))
	b.WriteString("\n")
	b.WriteString("Select the dependency folders you want to monitor:\n")
	b.WriteString("(space: toggle, a: toggle all, enter: confirm)\n\n")

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = wizardCursorStyle.Render("> ")
		}

		check := "[ ]"
		style := wizardItemStyle
		if item.selected {
			check = "[x]"
			style = wizardSelectedStyle
		}

		line := fmt.Sprintf("%s %s %s (%s)",
			cursor, check, item.target.Name, item.target.DependencyDir)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	selected := 0
	for _, item := range m.items {
		if item.selected {
			selected++
		}
	}
	b.WriteString(fmt.Sprintf("  %d selected  |  enter: confirm  |  q/esc: quit\n", selected))

	return b.String()
}
