package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// Snapshot captures the full state of all checks at a point in time.
type Snapshot struct {
	CapturedAt time.Time             `json:"captured_at"`
	Results    []check.Result        `json:"results"`
	Meta       map[string]string     `json:"meta,omitempty"`
}

// NewSnapshot creates a Snapshot from a slice of check results.
func NewSnapshot(results []check.Result) Snapshot {
	return Snapshot{
		CapturedAt: time.Now().UTC(),
		Results:    results,
		Meta:       make(map[string]string),
	}
}

// SaveSnapshot writes a snapshot as a JSON file under dir.
// The filename encodes the capture timestamp for easy ordering.
func SaveSnapshot(dir string, snap Snapshot) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("snapshot: create dir: %w", err)
	}
	name := fmt.Sprintf("snapshot_%s.json", snap.CapturedAt.Format("20060102T150405Z"))
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("snapshot: create file: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		return fmt.Errorf("snapshot: encode: %w", err)
	}
	return nil
}

// LoadSnapshot reads a single snapshot file from path.
func LoadSnapshot(path string) (Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: open: %w", err)
	}
	defer f.Close()
	var snap Snapshot
	if err := json.NewDecoder(f).Decode(&snap); err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: decode: %w", err)
	}
	return snap, nil
}
