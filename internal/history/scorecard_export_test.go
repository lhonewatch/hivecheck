package history

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/hivecheck/internal/check"
)

func TestExportScorecardCSV_NilWriter(t *testing.T) {
	err := ExportScorecardCSV(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil writer")
	}
}

func TestExportScorecardCSV_Header(t *testing.T) {
	var buf bytes.Buffer
	err := ExportScorecardCSV(nil, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	line := buf.String()
	for _, col := range []string{"check", "score", "availability_pct", "reliability_pct", "perf_pct", "last_status", "samples"} {
		if !strings.Contains(line, col) {
			t.Errorf("header missing column %q", col)
		}
	}
}

func TestExportScorecardCSV_Rows(t *testing.T) {
	entries := []ScorecardEntry{
		{
			CheckName:    "web",
			Score:        92.5,
			Availability: 95.0,
			Reliability:  98.0,
			Perf:         80.0,
			LastStatus:   check.StatusOK,
			Samples:      50,
		},
		{
			CheckName:    "db",
			Score:        60.0,
			Availability: 70.0,
			Reliability:  75.0,
			Perf:         25.0,
			LastStatus:   check.StatusWarn,
			Samples:      30,
		},
	}
	var buf bytes.Buffer
	if err := ExportScorecardCSV(entries, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// header + 2 data rows
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[1], "web") {
		t.Errorf("row 1 missing check name 'web': %s", lines[1])
	}
	if !strings.Contains(lines[2], "db") {
		t.Errorf("row 2 missing check name 'db': %s", lines[2])
	}
	if !strings.Contains(lines[1], "92.50") {
		t.Errorf("row 1 missing score 92.50: %s", lines[1])
	}
	if !strings.Contains(lines[2], "30") {
		t.Errorf("row 2 missing samples 30: %s", lines[2])
	}
}
