// Package filter provides staleness detection for project config files.
package filter

import (
	"os"
	"time"
)

// IsStale checks whether the file at path has a modification time older than
// the given number of days. Returns the stale status, the file's mod time,
// and any error encountered.
func IsStale(path string, days int) (bool, time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, time.Time{}, err
	}

	modTime := info.ModTime()
	threshold := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	return !modTime.After(threshold), modTime, nil
}
