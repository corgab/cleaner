package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot determine home directory: %v\n", err)
		os.Exit(1)
	}

	scanPath := flag.String("path", home, "Root directory to scan")
	days := flag.Int("days", 30, "Staleness threshold in days")
	dryRun := flag.Bool("dry-run", false, "Simulate without deleting")
	showStats := flag.Bool("stats", false, "Show lifetime statistics and exit")
	resetConfig := flag.Bool("reset-config", false, "Re-run the first-run wizard")
	flag.Parse()

	absPath, err := filepath.Abs(*scanPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid path %q: %v\n", *scanPath, err)
		os.Exit(1)
	}

	// Scaffold: print parsed flags for verification, replaced in Task 11
	fmt.Printf("Goclean\n  path: %s\n  days: %d\n  dry-run: %v\n  stats: %v\n  reset-config: %v\n",
		absPath, *days, *dryRun, *showStats, *resetConfig)
}
