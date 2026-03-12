// Package filter fornisce la logica di rilevamento della "stagnazione" dei progetti.
// Un progetto è considerato stale (inattivo) se il suo file di configurazione
// non è stato modificato da più di N giorni.
package filter

import (
	"os"
	"time"
)

// IsStale verifica se il file al percorso indicato ha una data di modifica
// più vecchia del numero di giorni specificato.
// Restituisce: stato di stale, data di modifica del file, eventuale errore.
func IsStale(path string, days int) (bool, time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, time.Time{}, err
	}

	modTime := info.ModTime()
	// Calcola la soglia: ora corrente meno N giorni
	threshold := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	// Il file è stale se la sua modifica NON è successiva alla soglia
	return !modTime.After(threshold), modTime, nil
}
