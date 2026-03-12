// Package cleaner handles removal of dependency directories.
package cleaner

import (
	"os"

	"github.com/corgab/goclean/internal/fsutil"
)

// Remove deletes the directory at path and returns the number of bytes freed.
// If dryRun is true, calculates size but does not delete.
func Remove(path string, dryRun bool) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return 0, os.ErrInvalid
	}

	size := fsutil.DirSize(path)

	if !dryRun {
		if err := os.RemoveAll(path); err != nil {
			return 0, err
		}
	}

	return size, nil
}
