package report_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
	"github.com/example/hivecheck/internal/report"
)

func makeSummary() *report.Summary {
	return &report.Summary{
		Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		Duration:  42 * time.Millisecond,
		Results: []check.Result{
			{Name: "db", Status: check.StatusOK, Message: "latency 5ms", Value: 5},
			{Name: "cache", Status: check.StatusWarn, Message: "latency 250ms", Value: 250},
			{Name: "queue", Status: check.StatusCritical, Message: "backlog high", Value: 9000},
		},
	}
}

func TestSummaryOverall(t *testing.T) {
	s := makeSummary()
	if got := s.Overall(); got != check.StatusCritical {
		t.Errorf("expected Critical, got %s", got)
	}
}

// TestSummaryOverall_AllOK verifies that Overall returns OK when all results are healthy.
func TestSummaryOverall_AllOK(t *testing.T) {
	s := &report.Summary{
		Timestamp: time.Now(),
		Duration:  1 * time.Millisecond,
		Results: []check.Result{
			{Name: "db", Status: check.StatusOK, Message: "ok", Value: 1},
			{Name: "cache", Status: check.StatusOK, Message: "ok", Value: 2},
		},
	}
	if got := s.Overall(); got != check.StatusOK {
		t.Errorf("expected OK, got %s", got)
	}
}

func TestSummaryStatusCounts(t *testing.T) {
	s := makeSummary()
	counts := s.StatusCounts()
	if counts[check.StatusOK] != 1 || counts[check.StatusWarn] != 1 || counts[check.StatusCritical] != 1 {
		t.Errorf("unexpected counts: %v", counts)
	}
}

func TestTextFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &report.TextFormatter{}
	if err := f.Format(&buf, makeSummary()); err != nil {
		t.Fatalf("Format error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"HiveCheck Report", "CRIT", "db", "cache", "queue", "OK=1", "WARN=1"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestJSONFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &report.JSONFormatter{Indent: true}
	if err := f.Format(&buf, makeSummary()); err != nil {
		t.Fatalf("Format error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{`"overall"`, `"CRITICAL"`, `"db"`, `"duration_ms"`} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON output missing %q", want)
		}
	}
}

func TestJSONFormatter_NoIndent(t *testing.T) {
	var buf bytes.Buffer
	f := &report.JSONFormatter{Indent: false}
	if err := f.Format(&buf, makeSummary()); err != nil {
		t.Fatalf("Format error: %v", err)
	}
	// compact output should not contain newlines beyond the trailing one
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("expected single-line JSON, got %d lines", len(lines))
	}
}
