// Package scanner walks the filesystem to find stale dependency directories.
package scanner

import (
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/corgab/goclean/internal/filter"
	"github.com/corgab/goclean/internal/fsutil"
	"github.com/corgab/goclean/pkg/targets"
)

// Options configures a scan.
type Options struct {
	Root          string
	Days          int
	Targets       []string // Enabled dependency dir names
	ExcludedPaths []string
}

// Result represents a found dependency directory.
type Result struct {
	Path          string    // Full path to the dependency directory
	ProjectDir    string    // Parent directory (the project root)
	DependencyDir string    // Directory name (e.g. "node_modules")
	TargetName    string    // Human-readable target name (e.g. "Node.js")
	ConfigFile    string    // Config file name that was matched
	Size          int64     // Total size in bytes
	ModTime       time.Time // Mod time of the config file
	Stale         bool
}

// systemPaths are always skipped during scanning.
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

// Scan walks the filesystem from opts.Root and returns all stale dependency
// directories matching the enabled targets.
func Scan(opts Options) ([]Result, error) {
	targetSet := make(map[string]bool, len(opts.Targets))
	for _, t := range opts.Targets {
		targetSet[t] = true
	}

	excludedSet := make(map[string]bool, len(opts.ExcludedPaths))
	for _, p := range opts.ExcludedPaths {
		excludedSet[p] = true
	}

	type found struct {
		path      string
		parentDir string
		dirName   string
	}

	var findings []found

	err := filepath.WalkDir(opts.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		if !d.IsDir() {
			return nil
		}

		name := d.Name()

		// Skip dot-directories (but not the root itself)
		if strings.HasPrefix(name, ".") && path != opts.Root {
			return fs.SkipDir
		}

		// Skip system directories at the root level
		if systemPaths[name] && filepath.Dir(path) == opts.Root {
			return fs.SkipDir
		}

		// Skip excluded paths
		if excludedSet[path] {
			return fs.SkipDir
		}

		// Check if this dir is a target dependency directory
		if targetSet[name] {
			findings = append(findings, found{
				path:      path,
				parentDir: filepath.Dir(path),
				dirName:   name,
			})
			return fs.SkipDir
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Process findings in parallel with a bounded worker pool
	results := make([]Result, len(findings))
	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU())

	for i, f := range findings {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, fd found) {
			defer wg.Done()
			defer func() { <-sem }()

			r := Result{
				Path:          fd.path,
				ProjectDir:    fd.parentDir,
				DependencyDir: fd.dirName,
			}

			r.Size = fsutil.DirSize(fd.path)

			// Disambiguate: check all targets for this dir name
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
					break
				}
			}
			if !matched {
				// No config file found in parent directory — skip this.
				// A "vendor" or "target" dir without its config file
				// (composer.json, go.mod, etc.) is almost certainly not
				// a dependency directory (e.g. Laravel's resources/views/vendor).
				return
			}

			results[idx] = r
		}(i, f)
	}

	wg.Wait()

	var staleResults []Result
	for _, r := range results {
		if r.Stale && r.Path != "" {
			staleResults = append(staleResults, r)
		}
	}

	return staleResults, nil
}
