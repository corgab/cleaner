package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/corgab/goclean/internal/config"
	"github.com/corgab/goclean/internal/scanner"
	"github.com/corgab/goclean/internal/stats"
	"github.com/corgab/goclean/internal/tui"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot determine home directory: %v\n", err)
		os.Exit(1)
	}

	scanPath := flag.String("path", home, "Root directory to scan")
	days := flag.Int("days", -1, "Staleness threshold in days (default: from config, typically 30)")
	dryRun := flag.Bool("dry-run", false, "Simulate without deleting")
	showStats := flag.Bool("stats", false, "Show lifetime statistics and exit")
	resetConfig := flag.Bool("reset-config", false, "Re-run the first-run wizard")
	flag.Parse()

	absPath, err := filepath.Abs(*scanPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid path %q: %v\n", *scanPath, err)
		os.Exit(1)
	}

	statsPath := filepath.Join(home, ".goclean_stats.json")
	configPath := filepath.Join(home, ".goclean.yaml")

	// Handle --stats
	if *showStats {
		s, err := stats.Load(statsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading stats: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(s.Summary())
		return
	}

	// First-run wizard or --reset-config
	if !config.Exists(configPath) || *resetConfig {
		wizard := tui.NewWizardModel()
		p := tea.NewProgram(wizard)
		finalModel, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		wm := finalModel.(tui.WizardModel)
		if wm.Quitting() {
			return
		}
		selected := wm.WizardResult()
		if len(selected) == 0 {
			fmt.Println("No targets selected. Exiting.")
			return
		}
		cfg := config.Default()
		cfg.Targets = selected
		if err := config.Save(cfg, configPath); err != nil {
			fmt.Fprintf(os.Stderr, "error saving config: %v\n", err)
			os.Exit(1)
		}
	}

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	// Apply flag overrides
	scanDays := cfg.Days
	if *days >= 0 {
		scanDays = *days
	}

	// Scan
	fmt.Printf("Scanning %s for stale dependencies (older than %d days)...\n", absPath, scanDays)
	results, err := scanner.Scan(scanner.Options{
		Root:          absPath,
		Days:          scanDays,
		Targets:       cfg.Targets,
		ExcludedPaths: cfg.ExcludedPaths,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
		os.Exit(1)
	}

	// Load stats for display
	s, _ := stats.Load(statsPath)

	// Run main TUI
	mainModel := tui.NewMainModel(results, *dryRun, s.Summary())
	p := tea.NewProgram(mainModel)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	mm := finalModel.(tui.MainModel)

	// Update stats if anything was deleted
	if mm.DeletedCount() > 0 {
		s.Record(mm.FreedBytes(), mm.DeletedCount())
		if err := stats.Save(s, statsPath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not save stats: %v\n", err)
		}
	}
}
