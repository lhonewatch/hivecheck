package history

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"
)

// ExportRollupCSV writes rollup buckets to w in CSV format.
// Columns: period_start, check_name, total, ok, warn, crit, error, worst_status, avg_duration_ms
func ExportRollupCSV(w io.Writer, buckets []RollupBucket) error {
	if w == nil {
		return fmt.Errorf("rollup export: writer must not be nil")
	}

	cw := csv.NewWriter(w)
	header := []string{
		"period_start",
		"check_name",
		"total",
		"ok",
		"warn",
		"crit",
		"error",
		"worst_status",
		"avg_duration_ms",
	}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("rollup export: write header: %w", err)
	}

	for _, b := range buckets {
		row := []string{
			b.PeriodStart.UTC().Format(time.RFC3339),
			b.CheckName,
			fmt.Sprintf("%d", b.Total),
			fmt.Sprintf("%d", b.OKCount),
			fmt.Sprintf("%d", b.WarnCount),
			fmt.Sprintf("%d", b.CritCount),
			fmt.Sprintf("%d", b.ErrorCount),
			b.WorstStatus().String(),
			fmt.Sprintf("%d", b.AvgDuration.Milliseconds()),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("rollup export: write row: %w", err)
		}
	}

	cw.Flush()
	return cw.Error()
}
