package stats_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/corgab/cleaner/internal/stats"
)

func TestNewStatsEmpty(t *testing.T) {
	s := stats.New()
	if s.TotalCleanedBytes != 0 {
		t.Errorf("expected 0 total bytes, got %d", s.TotalCleanedBytes)
	}
	if s.TotalDeletions != 0 {
		t.Errorf("expected 0 deletions, got %d", s.TotalDeletions)
	}
	if len(s.History) != 0 {
		t.Errorf("expected empty history, got %d entries", len(s.History))
	}
}

func TestRecordAndAccumulate(t *testing.T) {
	s := stats.New()
	s.Record(1024, 2)
	s.Record(2048, 3)

	if s.TotalCleanedBytes != 3072 {
		t.Errorf("expected 3072 total bytes, got %d", s.TotalCleanedBytes)
	}
	if s.TotalDeletions != 5 {
		t.Errorf("expected 5 deletions, got %d", s.TotalDeletions)
	}
	if len(s.History) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(s.History))
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stats.json")

	s := stats.New()
	s.Record(5000, 3)

	if err := stats.Save(&s, path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := stats.Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.TotalCleanedBytes != 5000 {
		t.Errorf("expected 5000 bytes, got %d", loaded.TotalCleanedBytes)
	}
	if loaded.TotalDeletions != 3 {
		t.Errorf("expected 3 deletions, got %d", loaded.TotalDeletions)
	}
	if loaded.FirstRun.IsZero() {
		t.Error("expected FirstRun to be set")
	}
}

func TestLoadMissingReturnsNew(t *testing.T) {
	s, err := stats.Load("/nonexistent/stats.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if s.TotalCleanedBytes != 0 {
		t.Error("expected fresh stats for missing file")
	}
}

func TestFormatSummary(t *testing.T) {
	s := stats.New()
	s.FirstRun = time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	s.TotalCleanedBytes = 15 * 1024 * 1024 * 1024
	s.TotalDeletions = 47

	summary := s.Summary()
	if summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestSummaryEmpty(t *testing.T) {
	s := stats.New()
	summary := s.Summary()
	if summary == "" {
		t.Error("expected non-empty summary even with no sessions")
	}
}
