// Package stats gestisce le statistiche cumulative delle sessioni di pulizia.
// I dati vengono salvati in formato JSON su ~/.corgab_stats.json.
package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/corgab/cleaner/internal/fsutil"
)

// Stats contiene le statistiche cumulative di tutte le sessioni di pulizia.
type Stats struct {
	TotalCleanedBytes int64          `json:"total_cleaned_bytes"` // Byte totali liberati
	TotalDeletions    int            `json:"total_deletions"`     // Numero totale di cartelle eliminate
	FirstRun          time.Time      `json:"first_run"`           // Data del primo utilizzo
	LastRun           time.Time      `json:"last_run"`            // Data dell'ultimo utilizzo
	History           []HistoryEntry `json:"history"`              // Storico delle singole sessioni
}

// HistoryEntry rappresenta una singola sessione di pulizia.
type HistoryEntry struct {
	Date       time.Time `json:"date"`        // Data della sessione
	FreedBytes int64     `json:"freed_bytes"` // Byte liberati nella sessione
	Count      int       `json:"count"`       // Cartelle eliminate nella sessione
}

// New restituisce un oggetto Stats vuoto e inizializzato.
func New() Stats {
	return Stats{
		History: []HistoryEntry{},
	}
}

// Record registra una nuova sessione di pulizia nelle statistiche.
// Aggiorna automaticamente FirstRun (se è il primo utilizzo) e LastRun.
func (s *Stats) Record(freedBytes int64, count int) {
	now := time.Now()
	if s.FirstRun.IsZero() {
		s.FirstRun = now
	}
	s.LastRun = now
	s.TotalCleanedBytes += freedBytes
	s.TotalDeletions += count
	s.History = append(s.History, HistoryEntry{
		Date:       now,
		FreedBytes: freedBytes,
		Count:      count,
	})
}

// Summary restituisce un riepilogo leggibile delle statistiche totali.
func (s *Stats) Summary() string {
	if s.TotalDeletions == 0 {
		return "No cleaning sessions yet."
	}
	return fmt.Sprintf("corgab cleaner has freed %s in %d operations since %s",
		fsutil.FormatBytes(s.TotalCleanedBytes),
		s.TotalDeletions,
		s.FirstRun.Format("02 Jan 2006"),
	)
}

// Save scrive le statistiche in formato JSON al percorso indicato.
func Save(s *Stats, path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Load legge le statistiche dal file JSON al percorso indicato.
// Se il file non esiste, restituisce statistiche vuote (primo avvio).
func Load(path string) (*Stats, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			s := New()
			return &s, nil
		}
		return nil, err
	}
	var s Stats
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
