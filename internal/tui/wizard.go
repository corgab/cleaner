// Package tui implementa l'interfaccia terminale interattiva basata su
// Bubbletea per Corgab. Contiene sia il wizard di primo avvio che la vista
// principale per la selezione e l'eliminazione delle cartelle.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/corgab/cleaner/pkg/targets"
)

// Stili Lipgloss per il wizard di configurazione
var (
	wizardTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
	wizardItemStyle     = lipgloss.NewStyle().PaddingLeft(2)
	wizardSelectedStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("120"))
	wizardCursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

// WizardModel è il modello Bubbletea per il wizard di selezione target al primo avvio.
// L'utente può selezionare quali cartelle di dipendenze monitorare.
type WizardModel struct {
	items    []wizardItem // Lista di target disponibili con stato di selezione
	cursor   int          // Posizione corrente del cursore
	done     bool         // true quando l'utente conferma con Enter
	quitting bool         // true quando l'utente esce con q/Esc
}

// wizardItem associa un target al suo stato di selezione nel wizard.
type wizardItem struct {
	target   targets.Target
	selected bool
}

// WizardResult restituisce i nomi delle cartelle di dipendenze selezionate dall'utente.
func (m WizardModel) WizardResult() []string {
	var selected []string
	for _, item := range m.items {
		if item.selected {
			selected = append(selected, item.target.DependencyDir)
		}
	}
	return selected
}

// Done indica se il wizard è stato completato con conferma.
func (m WizardModel) Done() bool { return m.done }

// Quitting indica se l'utente ha interrotto il wizard.
func (m WizardModel) Quitting() bool { return m.quitting }

// NewWizardModel crea un nuovo modello wizard con tutti i target disponibili.
func NewWizardModel() WizardModel {
	all := targets.All()
	items := make([]wizardItem, len(all))
	for i, t := range all {
		items[i] = wizardItem{target: t}
	}
	return WizardModel{items: items}
}

// Init è il comando iniziale di Bubbletea (nessuna azione richiesta).
func (m WizardModel) Init() tea.Cmd { return nil }

// Update gestisce tutti gli input da tastiera del wizard.
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
			// Spacebar: toggle selezione dell'elemento corrente
			m.items[m.cursor].selected = !m.items[m.cursor].selected
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case " ": // Spazio come rune (Windows compatibility)
				m.items[m.cursor].selected = !m.items[m.cursor].selected
			case "q":
				m.quitting = true
				return m, tea.Quit
			case "k": // Navigazione vim: su
				if m.cursor > 0 {
					m.cursor--
				}
			case "j": // Navigazione vim: giù
				if m.cursor < len(m.items)-1 {
					m.cursor++
				}
			case "a": // Toggle tutti: se tutti selezionati -> deseleziona, altrimenti seleziona tutti
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

// View renderizza l'interfaccia del wizard nel terminale.
func (m WizardModel) View() string {
	var b strings.Builder

	b.WriteString(wizardTitleStyle.Render("corgab cleaner - Setup Wizard"))
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
	b.WriteString(wizardItemStyle.Render("\n  github.com/corgab/cleaner"))
	b.WriteString("\n")

	return b.String()
}
