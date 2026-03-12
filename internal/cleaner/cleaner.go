// Package cleaner gestisce la rimozione delle cartelle di dipendenze dal disco.
package cleaner

import (
	"os"

	"github.com/corgab/cleaner/internal/fsutil"
)

// Remove elimina la directory al percorso indicato e restituisce i byte liberati.
// Se dryRun è true, calcola la dimensione senza eliminare nulla (simulazione).
func Remove(path string, dryRun bool) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	// Sicurezza: accetta solo directory, mai file singoli
	if !info.IsDir() {
		return 0, os.ErrInvalid
	}

	// Calcola la dimensione totale prima della rimozione
	size := fsutil.DirSize(path)

	if !dryRun {
		if err := os.RemoveAll(path); err != nil {
			return 0, err
		}
	}

	return size, nil
}
