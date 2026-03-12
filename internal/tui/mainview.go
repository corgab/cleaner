package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/corgab/goclean/internal/cleaner"
	"github.com/corgab/goclean/internal/fsutil"
	"github.com/corgab/goclean/internal/scanner"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)

	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("240"))

	itemNormalStyle   = lipgloss.NewStyle().PaddingLeft(2)
	itemSelectedStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("120"))
	cursorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	footerStyle = lipgloss.NewStyle().MarginTop(1).Foreground(lipgloss.Color("241"))

	confirmTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))

	successStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("120"))
)

type viewState int

const (
	stateList viewState = iota
	stateConfirm
	stateDeleting
	stateDone
)

// MainModel is the Bubbletea model for the main project list view.
type MainModel struct {
	results      []scanner.Result
	selected     []bool
	cursor       int
	state        viewState
	dryRun       bool
	statsSummary string

	deletedCount int
	freedBytes   int64
	deleteErrors []string
}

// DeletedCount returns how many directories were deleted.
func (m MainModel) DeletedCount() int { return m.deletedCount }

// FreedBytes returns the total bytes freed.
func (m MainModel) FreedBytes() int64 { return m.freedBytes }

// SelectedPaths returns the paths the user selected for deletion.
func (m MainModel) SelectedPaths() []string {
	var paths []string
	for i, sel := range m.selected {
		if sel {
			paths = append(paths, m.results[i].Path)
		}
	}
	return paths
}

// NewMainModel creates the main view model.
func NewMainModel(results []scanner.Result, dryRun bool, statsSummary string) MainModel {
	return MainModel{
		results:      results,
		selected:     make([]bool, len(results)),
		dryRun:       dryRun,
		statsSummary: statsSummary,
	}
}

func (m MainModel) Init() tea.Cmd { return nil }

type deletionDoneMsg struct {
	count  int
	freed  int64
	errors []string
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case deletionDoneMsg:
		m.deletedCount = msg.count
		m.freedBytes = msg.freed
		m.deleteErrors = msg.errors
		m.state = stateDone
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateList:
			return m.updateList(msg)
		case stateConfirm:
			return m.updateConfirm(msg)
		case stateDone:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MainModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		return m, tea.Quit
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEnter:
		if m.selectedCount() > 0 {
			m.state = stateConfirm
		}
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < len(m.results)-1 {
			m.cursor++
		}
	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "q":
			return m, tea.Quit
		case "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "j":
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
		case " ":
			if len(m.results) > 0 {
				m.selected[m.cursor] = !m.selected[m.cursor]
			}
		case "a":
			allSel := true
			for _, s := range m.selected {
				if !s {
					allSel = false
					break
				}
			}
			for i := range m.selected {
				m.selected[i] = !allSel
			}
		case "n":
			for i := range m.selected {
				m.selected[i] = false
			}
		}
	}
	return m, nil
}

func (m MainModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.state = stateList
		return m, nil
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "y", "Y":
			m.state = stateDeleting
			return m, m.performDeletion()
		case "n", "N":
			m.state = stateList
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MainModel) performDeletion() tea.Cmd {
	return func() tea.Msg {
		var (
			totalFreed int64
			count      int
			errors     []string
		)
		for i, sel := range m.selected {
			if !sel {
				continue
			}
			r := m.results[i]
			freed, err := cleaner.Remove(r.Path, m.dryRun)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", r.Path, err))
			} else {
				totalFreed += freed
				count++
			}
		}
		return deletionDoneMsg{count: count, freed: totalFreed, errors: errors}
	}
}

func (m MainModel) View() string {
	switch m.state {
	case stateList:
		return m.viewList()
	case stateConfirm:
		return m.viewConfirm()
	case stateDeleting:
		return m.viewDeleting()
	case stateDone:
		return m.viewDone()
	}
	return ""
}

func (m MainModel) viewList() string {
	var b strings.Builder

	title := "Goclean"
	if m.dryRun {
		title += " (DRY RUN)"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	if len(m.results) == 0 {
		b.WriteString("  No stale dependency folders found. Your disk is clean!\n")
		b.WriteString(footerStyle.Render("\n  q: quit"))
		return b.String()
	}

	b.WriteString(headerStyle.Render(fmt.Sprintf("  %-50s %-14s %-10s %s", "Project", "Type", "Size", "Last Modified")))
	b.WriteString("\n")

	for i, r := range m.results {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}

		check := "[ ]"
		style := itemNormalStyle
		if m.selected[i] {
			check = "[x]"
			style = itemSelectedStyle
		}

		modStr := "N/A"
		if !r.ModTime.IsZero() {
			modStr = r.ModTime.Format("02 Jan 2006")
		}
		if r.MissingConfig {
			modStr = "no config"
		}

		line := fmt.Sprintf("%s %s %-45s %-14s %-10s %s",
			cursor, check,
			truncate(r.ProjectDir, 45),
			r.TargetName,
			fsutil.FormatBytes(r.Size),
			modStr,
		)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(footerStyle.Render(fmt.Sprintf(
		"  Selected: %d/%d  |  Space to free: %s",
		m.selectedCount(), len(m.results), fsutil.FormatBytes(m.selectedSize()),
	)))
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("  space: toggle  |  a: all  |  n: none  |  enter: delete  |  q: quit"))

	if m.statsSummary != "" {
		b.WriteString("\n")
		b.WriteString(footerStyle.Render("  " + m.statsSummary))
	}

	return b.String()
}

func (m MainModel) viewConfirm() string {
	var b strings.Builder

	b.WriteString(confirmTitleStyle.Render("Confirm Deletion"))
	b.WriteString("\n\n")

	if m.dryRun {
		b.WriteString("  This is a DRY RUN - nothing will be deleted.\n\n")
	}

	for i, sel := range m.selected {
		if sel {
			r := m.results[i]
			b.WriteString(fmt.Sprintf("  - %s (%s)\n", r.Path, fsutil.FormatBytes(r.Size)))
		}
	}

	b.WriteString(fmt.Sprintf("\n  Total: %s in %d folders\n",
		fsutil.FormatBytes(m.selectedSize()), m.selectedCount()))
	b.WriteString("\n  Press y to confirm, n to go back\n")

	return b.String()
}

func (m MainModel) viewDeleting() string {
	if m.dryRun {
		return "\n  Simulating deletion...\n"
	}
	return fmt.Sprintf("\n  Deleting %d folders...\n", m.selectedCount())
}

func (m MainModel) viewDone() string {
	var b strings.Builder

	if m.dryRun {
		b.WriteString(successStyle.Render("Dry run complete"))
	} else {
		b.WriteString(successStyle.Render("Deletion complete"))
	}
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  Freed: %s from %d folders\n", fsutil.FormatBytes(m.freedBytes), m.deletedCount))

	if len(m.deleteErrors) > 0 {
		b.WriteString(fmt.Sprintf("\n  Errors (%d):\n", len(m.deleteErrors)))
		for _, e := range m.deleteErrors {
			b.WriteString(fmt.Sprintf("    - %s\n", e))
		}
	}

	b.WriteString("\n  Press any key to exit\n")
	return b.String()
}

func (m MainModel) selectedCount() int {
	count := 0
	for _, s := range m.selected {
		if s {
			count++
		}
	}
	return count
}

func (m MainModel) selectedSize() int64 {
	var total int64
	for i, s := range m.selected {
		if s {
			total += m.results[i].Size
		}
	}
	return total
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return "..." + s[len(s)-maxLen+3:]
}
