package history

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"
)

// ExportCSV writes all stored entries for the given check names to w in CSV
// format. If names is empty, entries for every check found in the store are
// exported. The CSV header row is always written first.
func ExportCSV(store *Store, w io.Writer, names []string) error {
	if store == nil {
		return fmt.Errorf("history: store must not be nil")
	}

	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"check", "timestamp", "status", "message", "duration_ms"}); err != nil {
		return fmt.Errorf("history: writing csv header: %w", err)
	}

	targets := names
	if len(targets) == 0 {
		all, err := store.CheckNames()
		if err != nil {
			return fmt.Errorf("history: listing check names: %w", err)
		}
		targets = all
	}

	for _, name := range targets {
		entries, err := store.Load(name)
		if err != nil {
			return fmt.Errorf("history: loading entries for %q: %w", name, err)
		}
		for _, e := range entries {
			row := []string{
				e.CheckName,
				e.Timestamp.UTC().Format(time.RFC3339),
				e.Status,
				e.Message,
				fmt.Sprintf("%d", e.DurationMs),
			}
			if err := cw.Write(row); err != nil {
				return fmt.Errorf("history: writing csv row: %w", err)
			}
		}
	}

	cw.Flush()
	return cw.Error()
}
