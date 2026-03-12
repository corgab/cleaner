// Package stats tracks lifetime cleaning statistics for Goclean.
package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/corgab/goclean/internal/fsutil"
)

// Stats holds lifetime cleaning statistics.
type Stats struct {
	TotalCleanedBytes int64          `json:"total_cleaned_bytes"`
	TotalDeletions    int            `json:"total_deletions"`
	FirstRun          time.Time      `json:"first_run"`
	LastRun           time.Time      `json:"last_run"`
	History           []HistoryEntry `json:"history"`
}

// HistoryEntry records a single cleaning session.
type HistoryEntry struct {
	Date       time.Time `json:"date"`
	FreedBytes int64     `json:"freed_bytes"`
	Count      int       `json:"count"`
}

// New returns a fresh, empty Stats.
func New() Stats {
	return Stats{
		History: []HistoryEntry{},
	}
}

// Record adds a cleaning session to the stats.
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

// Summary returns a human-readable summary string.
func (s *Stats) Summary() string {
	if s.TotalDeletions == 0 {
		return "No cleaning sessions yet."
	}
	return fmt.Sprintf("Goclean has freed %s in %d operations since %s",
		fsutil.FormatBytes(s.TotalCleanedBytes),
		s.TotalDeletions,
		s.FirstRun.Format("02 Jan 2006"),
	)
}

// Save writes stats as JSON to the given path.
func Save(s *Stats, path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Load reads stats from the given path. Returns fresh stats if file is missing.
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
