// Punto di ingresso di Corgab: parsing dei flag, orchestrazione del wizard,
// della scansione e della TUI interattiva.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/corgab/cleaner/internal/config"
	"github.com/corgab/cleaner/internal/scanner"
	"github.com/corgab/cleaner/internal/stats"
	"github.com/corgab/cleaner/internal/tui"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "errore: impossibile determinare la home directory: %v\n", err)
		os.Exit(1)
	}

	// Definizione dei flag da linea di comando
	scanPath := flag.String("path", home, "Directory radice da scansionare")
	days := flag.Int("days", -1, "Soglia di giorni di inattività (default: dal config, tipicamente 30)")
	dryRun := flag.Bool("dry-run", false, "Simula l'eliminazione senza cancellare nulla")
	showStats := flag.Bool("stats", false, "Mostra le statistiche cumulative ed esci")
	resetConfig := flag.Bool("reset-config", false, "Riesegui il wizard di configurazione")
	flag.Parse()

	// Converte il percorso in assoluto per evitare ambiguità
	absPath, err := filepath.Abs(*scanPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "errore: percorso non valido %q: %v\n", *scanPath, err)
		os.Exit(1)
	}

	// Percorsi dei file di persistenza
	statsPath := filepath.Join(home, ".corgab_stats.json")
	configPath := filepath.Join(home, ".corgab.yaml")

	// Gestione flag --stats: mostra statistiche ed esci
	if *showStats {
		s, err := stats.Load(statsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "errore nel caricamento delle statistiche: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(s.Summary())
		return
	}

	// Wizard di primo avvio o --reset-config:
	// l'utente seleziona quali cartelle di dipendenze monitorare
	if !config.Exists(configPath) || *resetConfig {
		wizard := tui.NewWizardModel()
		p := tea.NewProgram(wizard)
		finalModel, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "errore: %v\n", err)
			os.Exit(1)
		}
		wm := finalModel.(tui.WizardModel)
		if wm.Quitting() {
			return
		}
		selected := wm.WizardResult()
		if len(selected) == 0 {
			fmt.Println("Nessun target selezionato. Uscita.")
			return
		}
		cfg := config.Default()
		cfg.Targets = selected
		if err := config.Save(cfg, configPath); err != nil {
			fmt.Fprintf(os.Stderr, "errore nel salvataggio della configurazione: %v\n", err)
			os.Exit(1)
		}
	}

	// Carica la configurazione salvata
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "errore nel caricamento della configurazione: %v\n", err)
		os.Exit(1)
	}

	// Il flag --days ha priorità sul valore nel config.
	// Il sentinel -1 indica "usa il valore dal config".
	scanDays := cfg.Days
	if *days >= 0 {
		scanDays = *days
	}

	// Avvia la scansione del filesystem
	fmt.Printf("Scanning %s for stale dependencies (older than %d days)...\n", absPath, scanDays)
	results, err := scanner.Scan(scanner.Options{
		Root:          absPath,
		Days:          scanDays,
		Targets:       cfg.Targets,
		ExcludedPaths: cfg.ExcludedPaths,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "errore durante la scansione: %v\n", err)
		os.Exit(1)
	}

	// Carica le statistiche per mostrarle nel footer della TUI
	s, _ := stats.Load(statsPath)

	// Avvia la TUI principale con i risultati della scansione
	mainModel := tui.NewMainModel(results, *dryRun, s.Summary())
	p := tea.NewProgram(mainModel)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "errore: %v\n", err)
		os.Exit(1)
	}
	mm := finalModel.(tui.MainModel)

	// Aggiorna e salva le statistiche se qualcosa è stato eliminato
	if mm.DeletedCount() > 0 {
		s.Record(mm.FreedBytes(), mm.DeletedCount())
		if err := stats.Save(s, statsPath); err != nil {
			fmt.Fprintf(os.Stderr, "attenzione: impossibile salvare le statistiche: %v\n", err)
		}
	}
}
