package history

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ListSnapshots returns the file paths of all snapshots in dir,
// sorted chronologically (oldest first) by filename.
func ListSnapshots(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("snapshot: read dir: %w", err)
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), "snapshot_") && strings.HasSuffix(e.Name(), ".json") {
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(paths)
	return paths, nil
}

// LatestSnapshot returns the most recently saved snapshot from dir.
// Returns ErrNoSnapshots if none exist.
func LatestSnapshot(dir string) (Snapshot, error) {
	paths, err := ListSnapshots(dir)
	if err != nil {
		return Snapshot{}, err
	}
	if len(paths) == 0 {
		return Snapshot{}, ErrNoSnapshots
	}
	return LoadSnapshot(paths[len(paths)-1])
}

// ErrNoSnapshots is returned when no snapshot files are found.
var ErrNoSnapshots = fmt.Errorf("snapshot: no snapshots found")
