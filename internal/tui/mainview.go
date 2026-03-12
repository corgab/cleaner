package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/corgab/cleaner/internal/cleaner"
	"github.com/corgab/cleaner/internal/fsutil"
	"github.com/corgab/cleaner/internal/scanner"
)

// Stili Lipgloss per la vista principale
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

// viewState rappresenta lo stato corrente della vista principale.
// La TUI è una macchina a stati: lista -> conferma -> eliminazione -> completato.
type viewState int

const (
	stateList     viewState = iota // Lista navigabile dei risultati
	stateConfirm                   // Schermata di conferma prima dell'eliminazione
	stateDeleting                  // Eliminazione in corso
	stateDone                      // Operazione completata
)

// MainModel è il modello Bubbletea per la vista principale con la lista dei progetti.
type MainModel struct {
	results      []scanner.Result // Risultati della scansione
	selected     []bool           // Stato di selezione per ogni risultato
	cursor       int              // Posizione corrente del cursore
	state        viewState        // Stato della macchina a stati della vista
	dryRun       bool             // Se true, simula l'eliminazione senza cancellare
	statsSummary string           // Riepilogo statistiche da mostrare nel footer

	deletedCount int      // Contatore cartelle eliminate
	freedBytes   int64    // Byte totali liberati
	deleteErrors []string // Errori verificatisi durante l'eliminazione
}

// DeletedCount restituisce il numero di cartelle eliminate.
func (m MainModel) DeletedCount() int { return m.deletedCount }

// FreedBytes restituisce il totale dei byte liberati.
func (m MainModel) FreedBytes() int64 { return m.freedBytes }

// SelectedPaths restituisce i percorsi selezionati dall'utente per l'eliminazione.
func (m MainModel) SelectedPaths() []string {
	var paths []string
	for i, sel := range m.selected {
		if sel {
			paths = append(paths, m.results[i].Path)
		}
	}
	return paths
}

// NewMainModel crea il modello della vista principale.
func NewMainModel(results []scanner.Result, dryRun bool, statsSummary string) MainModel {
	return MainModel{
		results:      results,
		selected:     make([]bool, len(results)),
		dryRun:       dryRun,
		statsSummary: statsSummary,
	}
}

// Init è il comando iniziale di Bubbletea (nessuna azione richiesta).
func (m MainModel) Init() tea.Cmd { return nil }

// deletionDoneMsg è il messaggio Bubbletea emesso al completamento dell'eliminazione.
type deletionDoneMsg struct {
	count  int
	freed  int64
	errors []string
}

// Update è il dispatcher principale degli eventi: smista i messaggi
// in base allo stato corrente della macchina a stati.
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case deletionDoneMsg:
		// Eliminazione completata: aggiorna i contatori e passa a stateDone
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
			return m, tea.Quit // Qualsiasi tasto per uscire
		}
	}
	return m, nil
}

// updateList gestisce gli input nella vista lista (navigazione e selezione).
func (m MainModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		return m, tea.Quit
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEnter:
		// Enter: procedi alla conferma solo se c'è almeno una selezione
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
	case tea.KeySpace:
		// Spacebar: toggle selezione dell'elemento corrente
		if len(m.results) > 0 {
			m.selected[m.cursor] = !m.selected[m.cursor]
		}
	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "q":
			return m, tea.Quit
		case "k": // Navigazione vim: su
			if m.cursor > 0 {
				m.cursor--
			}
		case "j": // Navigazione vim: giù
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
		case "a": // Seleziona/deseleziona tutti
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
		case "n": // Deseleziona tutti
			for i := range m.selected {
				m.selected[i] = false
			}
		}
	}
	return m, nil
}

// updateConfirm gestisce gli input nella schermata di conferma.
func (m MainModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.state = stateList // Torna alla lista
		return m, nil
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "y", "Y":
			// Conferma: avvia l'eliminazione in background
			m.state = stateDeleting
			return m, m.performDeletion()
		case "n", "N":
			m.state = stateList // Annulla, torna alla lista
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

// performDeletion restituisce un comando Bubbletea che esegue l'eliminazione
// delle cartelle selezionate in background. Al termine emette un deletionDoneMsg.
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

// View renderizza la vista corrente in base allo stato della macchina a stati.
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

// viewList renderizza la lista navigabile dei risultati di scansione.
func (m MainModel) viewList() string {
	var b strings.Builder

	title := "corgab cleaner"
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

	// Header della tabella
	b.WriteString(headerStyle.Render(fmt.Sprintf("  %-60s %-10s %s", "Folder to delete", "Size", "Last Modified")))
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

		// Mostra il percorso effettivo che verrà eliminato (es. .../myproject/vendor)
		displayPath := truncate(r.Path, 50)

		line := fmt.Sprintf("%s %s %-50s %-10s %s",
			cursor, check,
			displayPath,
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

	b.WriteString("\n")
	b.WriteString(footerStyle.Render("  github.com/corgab/cleaner"))

	return b.String()
}

// viewConfirm renderizza la schermata di conferma con l'elenco delle cartelle
// che stanno per essere eliminate e la dimensione totale.
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

// viewDeleting renderizza il messaggio durante l'eliminazione in corso.
func (m MainModel) viewDeleting() string {
	if m.dryRun {
		return "\n  Simulating deletion...\n"
	}
	return fmt.Sprintf("\n  Deleting %d folders...\n", m.selectedCount())
}

// viewDone renderizza il riepilogo finale dopo l'eliminazione.
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

// selectedCount restituisce il numero di elementi attualmente selezionati.
func (m MainModel) selectedCount() int {
	count := 0
	for _, s := range m.selected {
		if s {
			count++
		}
	}
	return count
}

// selectedSize restituisce la dimensione totale in byte degli elementi selezionati.
func (m MainModel) selectedSize() int64 {
	var total int64
	for i, s := range m.selected {
		if s {
			total += m.results[i].Size
		}
	}
	return total
}

// truncate tronca una stringa aggiungendo "..." all'inizio se supera maxLen.
// Utile per mostrare percorsi lunghi in spazi limitati.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return "..." + s[len(s)-maxLen+3:]
}
