// Package scanner percorre il filesystem per individuare le cartelle di
// dipendenze stale (inattive). Utilizza goroutine con pool limitato per
// massimizzare la velocità di scansione senza sovraccaricare il sistema.
package scanner

import (
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/corgab/cleaner/internal/filter"
	"github.com/corgab/cleaner/internal/fsutil"
	"github.com/corgab/cleaner/pkg/targets"
)

// Options configura i parametri di una scansione.
type Options struct {
	Root          string   // Directory radice da cui iniziare la scansione
	Days          int      // Soglia di giorni: mostra solo progetti più vecchi di N giorni
	Targets       []string // Nomi delle cartelle di dipendenze abilitate (es. "node_modules")
	ExcludedPaths []string // Percorsi da escludere completamente dalla scansione
}

// Result rappresenta una cartella di dipendenze trovata durante la scansione.
type Result struct {
	Path          string    // Percorso completo della cartella di dipendenze
	ProjectDir    string    // Directory genitore (root del progetto)
	DependencyDir string    // Nome della cartella (es. "node_modules")
	TargetName    string    // Nome leggibile del target (es. "Node.js")
	ConfigFile    string    // File di configurazione trovato nel progetto
	Size          int64     // Dimensione totale in byte
	ModTime       time.Time // Data di modifica del file di configurazione
	Stale         bool      // true se il progetto è considerato inattivo
}

// systemPaths contiene le directory di sistema da ignorare sempre.
// Sono le cartelle tipiche di macOS/Linux che non contengono progetti utente.
var systemPaths = map[string]bool{
	"System":  true,
	"Library": true,
	"usr":     true,
	"bin":     true,
	"sbin":    true,
	"proc":    true,
	"sys":     true,
	"dev":     true,
}

// Scan percorre il filesystem a partire da opts.Root e restituisce tutte le
// cartelle di dipendenze stale che corrispondono ai target abilitati.
//
// Il processo avviene in due fasi:
//  1. Walk sincrono del filesystem per trovare le cartelle candidate
//  2. Analisi parallela delle candidate (dimensione, verifica config, staleness)
func Scan(opts Options) ([]Result, error) {
	// Costruisce i set di lookup per accesso O(1)
	targetSet := make(map[string]bool, len(opts.Targets))
	for _, t := range opts.Targets {
		targetSet[t] = true
	}

	excludedSet := make(map[string]bool, len(opts.ExcludedPaths))
	for _, p := range opts.ExcludedPaths {
		excludedSet[p] = true
	}

	// Struttura temporanea per le cartelle trovate durante il walk
	type found struct {
		path      string
		parentDir string
		dirName   string
	}

	var findings []found

	// Fase 1: Walk sincrono del filesystem
	err := filepath.WalkDir(opts.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir // Ignora directory con errori di accesso
		}
		if !d.IsDir() {
			return nil // Ignora i file, cerchiamo solo directory
		}

		name := d.Name()

		// Salta le directory nascoste (iniziano con "."), esclusa la root
		if strings.HasPrefix(name, ".") && path != opts.Root {
			return fs.SkipDir
		}

		// Salta le directory di sistema al primo livello
		if systemPaths[name] && filepath.Dir(path) == opts.Root {
			return fs.SkipDir
		}

		// Salta i percorsi esclusi dall'utente
		if excludedSet[path] {
			return fs.SkipDir
		}

		// Se la directory corrisponde a un target abilitato, registrala
		if targetSet[name] {
			findings = append(findings, found{
				path:      path,
				parentDir: filepath.Dir(path),
				dirName:   name,
			})
			return fs.SkipDir // Non entrare dentro la cartella di dipendenze
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Fase 2: Analisi parallela con pool di goroutine limitato.
	// Il semaforo limita la concorrenza al numero di CPU disponibili
	// per evitare di esaurire i file descriptor del sistema operativo.
	results := make([]Result, len(findings))
	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU())

	for i, f := range findings {
		wg.Add(1)
		sem <- struct{}{} // Acquisisce uno slot dal semaforo
		go func(idx int, fd found) {
			defer wg.Done()
			defer func() { <-sem }() // Rilascia lo slot

			r := Result{
				Path:          fd.path,
				ProjectDir:    fd.parentDir,
				DependencyDir: fd.dirName,
			}

			// Calcola la dimensione totale della cartella
			r.Size = fsutil.DirSize(fd.path)

			// Disambiguazione: per cartelle condivise (es. "vendor"),
			// verifica tutti i possibili config file per identificare
			// l'ecosistema corretto (PHP vs Go, Rust vs Java, ecc.)
			allTargets := targets.GetAllByDirName(fd.dirName)
			matched := false
			for _, tgt := range allTargets {
				configPath := filepath.Join(fd.parentDir, tgt.ConfigFile)
				stale, modTime, err := filter.IsStale(configPath, opts.Days)
				if err == nil {
					r.TargetName = tgt.Name
					r.ConfigFile = tgt.ConfigFile
					r.Stale = stale
					r.ModTime = modTime
					matched = true
					break // Trovato il config file corretto, non serve continuare
				}
			}
			if !matched {
				// Nessun file di configurazione trovato nella directory genitore.
				// Questa cartella NON è una vera directory di dipendenze
				// (es. vendor/ dentro Laravel non è il vendor di Composer).
				// La ignoriamo per evitare falsi positivi.
				return
			}

			results[idx] = r
		}(i, f)
	}

	wg.Wait()

	// Filtra: restituisce solo i risultati stale con percorso valido
	var staleResults []Result
	for _, r := range results {
		if r.Stale && r.Path != "" {
			staleResults = append(staleResults, r)
		}
	}

	return staleResults, nil
}
