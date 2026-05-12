package history

import (
	"encoding/csv"
	"fmt"
	"io"
)

// ExportScorecardCSV writes scorecard entries to w in CSV format.
// The first row is a header. Returns an error if writing fails or entries is nil.
func ExportScorecardCSV(entries []ScorecardEntry, w io.Writer) error {
	if w == nil {
		return fmt.Errorf("scorecard export: writer must not be nil")
	}
	cw := csv.NewWriter(w)
	header := []string{"check", "score", "availability_pct", "reliability_pct", "perf_pct", "last_status", "samples"}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("scorecard export: write header: %w", err)
	}
	for _, e := range entries {
		row := []string{
			e.CheckName,
			fmt.Sprintf("%.2f", e.Score),
			fmt.Sprintf("%.2f", e.Availability),
			fmt.Sprintf("%.2f", e.Reliability),
			fmt.Sprintf("%.2f", e.Perf),
			e.LastStatus.String(),
			fmt.Sprintf("%d", e.Samples),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("scorecard export: write row: %w", err)
		}
	}
	cw.Flush()
	return cw.Error()
}
