package history

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
)

func TestExportRollupCSV_Header(t *testing.T) {
	var buf bytes.Buffer
	if err := ExportRollupCSV(&buf, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line (header only), got %d", len(lines))
	}
	for _, col := range []string{"period_start", "check_name", "total", "worst_status", "avg_duration_ms"} {
		if !strings.Contains(lines[0], col) {
			t.Errorf("header missing column %q", col)
		}
	}
}

func TestExportRollupCSV_Rows(t *testing.T) {
	buckets := []RollupBucket{
		{
			PeriodStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			CheckName:   "api",
			Total:       4,
			OKCount:     3,
			WarnCount:   1,
			AvgDuration: 120 * time.Millisecond,
		},
		{
			PeriodStart: time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC),
			CheckName:   "db",
			Total:       2,
			CritCount:   2,
			AvgDuration: 500 * time.Millisecond,
		},
	}

	var buf bytes.Buffer
	if err := ExportRollupCSV(&buf, buckets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "api") {
		t.Error("output missing 'api'")
	}
	if !strings.Contains(output, "db") {
		t.Error("output missing 'db'")
	}
	if !strings.Contains(output, check.StatusCritical.String()) {
		t.Errorf("output missing critical status")
	}
	if !strings.Contains(output, "120") {
		t.Error("output missing avg duration 120ms")
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines (header + 2 rows), got %d", len(lines))
	}
}

func TestExportRollupCSV_NilWriter(t *testing.T) {
	err := ExportRollupCSV(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil writer")
	}
}
