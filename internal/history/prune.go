package history

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// PruneOptions controls how old history entries are removed.
type PruneOptions struct {
	// MaxAge removes entries older than this duration. Zero means no age limit.
	MaxAge time.Duration
	// MaxEntries keeps only the N most-recent entries per check name.
	// Zero means no limit.
	MaxEntries int
}

// Prune removes history entries from dir according to opts.
// It returns the number of files deleted and any first error encountered.
func Prune(dir string, opts PruneOptions) (deleted int, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	// Group JSON files by check name (prefix before first '-').
	groups := make(map[string][]os.DirEntry)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		name := checkNameFromFile(e.Name())
		groups[name] = append(groups[name], e)
	}

	now := time.Now()

	for _, files := range groups {
		// Sort ascending by name (names embed timestamps, so lexicographic == chronological).
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name() < files[j].Name()
		})

		for idx, f := range files {
			path := filepath.Join(dir, f.Name())
			remove := false

			if opts.MaxEntries > 0 && len(files)-idx > opts.MaxEntries {
				remove = true
			}

			if !remove && opts.MaxAge > 0 {
				info, statErr := f.Info()
				if statErr == nil && now.Sub(info.ModTime()) > opts.MaxAge {
					remove = true
				}
			}

			if remove {
				if rmErr := os.Remove(path); rmErr != nil && err == nil {
					err = rmErr
				} else {
					deleted++
				}
			}
		}
	}

	return deleted, err
}

// checkNameFromFile extracts the logical check name from a history filename.
// Files are named "<checkName>-<timestamp>.json".
func checkNameFromFile(filename string) string {
	base := strings.TrimSuffix(filename, ".json")
	// Find the last '-' that separates name from timestamp.
	if idx := strings.LastIndex(base, "-"); idx > 0 {
		return base[:idx]
	}
	return base
}
