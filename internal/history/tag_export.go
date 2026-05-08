package history

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// ExportTaggedCSV writes tagged entries to w as CSV.
// Columns: check_name, status, duration_ms, timestamp, tags
func ExportTaggedCSV(w io.Writer, entries []TaggedEntry) error {
	if w == nil {
		return fmt.Errorf("history: ExportTaggedCSV: writer is nil")
	}
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"check_name", "status", "duration_ms", "timestamp", "tags"}); err != nil {
		return fmt.Errorf("history: ExportTaggedCSV: write header: %w", err)
	}
	for _, e := range entries {
		row := []string{
			e.CheckName,
			e.Status,
			fmt.Sprintf("%d", e.Duration.Milliseconds()),
			e.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
			strings.Join(e.Tags, "|"),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("history: ExportTaggedCSV: write row: %w", err)
		}
	}
	cw.Flush()
	return cw.Error()
}
