// Package history provides persistent storage for check run results,
// enabling trend analysis and comparison across multiple executions.
package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// Entry represents a single stored check run.
type Entry struct {
	Timestamp time.Time              `json:"timestamp"`
	Results   []check.Result         `json:"results"`
	Overall   check.Status           `json:"overall"`
}

// Store persists and retrieves historical check entries.
type Store struct {
	dir string
}

// NewStore creates a Store backed by the given directory.
// The directory is created if it does not exist.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("history: create dir %q: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

// Save persists an entry to disk using a timestamp-based filename.
func (s *Store) Save(e Entry) error {
	name := e.Timestamp.UTC().Format("20060102T150405Z") + ".json"
	path := filepath.Join(s.dir, name)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("history: create file: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(e); err != nil {
		return fmt.Errorf("history: encode entry: %w", err)
	}
	return nil
}

// Load reads all stored entries from disk, sorted by filename (chronological).
func (s *Store) Load() ([]Entry, error) {
	glob := filepath.Join(s.dir, "*.json")
	matches, err := filepath.Glob(glob)
	if err != nil {
		return nil, fmt.Errorf("history: glob: %w", err)
	}
	entries := make([]Entry, 0, len(matches))
	for _, m := range matches {
		e, err := readEntry(m)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func readEntry(path string) (Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return Entry{}, fmt.Errorf("history: open %q: %w", path, err)
	}
	defer f.Close()
	var e Entry
	if err := json.NewDecoder(f).Decode(&e); err != nil {
		return Entry{}, fmt.Errorf("history: decode %q: %w", path, err)
	}
	return e, nil
}
