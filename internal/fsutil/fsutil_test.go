package fsutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/corgab/goclean/internal/fsutil"
)

func TestDirSize(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), make([]byte, 1024), 0644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), make([]byte, 2048), 0644); err != nil {
		t.Fatal(err)
	}

	size := fsutil.DirSize(dir)
	if size < 3072 {
		t.Errorf("expected at least 3072 bytes, got %d", size)
	}
}

func TestDirSizeEmpty(t *testing.T) {
	dir := t.TempDir()
	size := fsutil.DirSize(dir)
	if size != 0 {
		t.Errorf("expected 0 bytes for empty dir, got %d", size)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1610612736, "1.5 GB"},
	}
	for _, tc := range tests {
		got := fsutil.FormatBytes(tc.input)
		if got != tc.expected {
			t.Errorf("FormatBytes(%d) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
