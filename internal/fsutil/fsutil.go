// Package fsutil fornisce utilità condivise per operazioni sul filesystem.
// Usato sia dallo scanner (per calcolare le dimensioni) che dal cleaner.
package fsutil

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

// DirSize calcola la dimensione totale in byte di tutti i file contenuti
// ricorsivamente in una directory. Gli errori di accesso vengono ignorati
// silenziosamente (es. permessi mancanti).
func DirSize(path string) int64 {
	var total int64
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Ignora errori di accesso e continua
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}

// FormatBytes converte un valore in byte in una stringa leggibile (B, KB, MB, GB).
func FormatBytes(b int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
