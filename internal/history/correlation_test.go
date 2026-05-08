package history

import (
	"strings"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
)

func buildCorrEntries(names []string, statuses [][]check.Status, base time.Time) []Entry {
	var entries []Entry
	for row, name := range names {
		for col, s := range statuses[row] {
			entries = append(entries, Entry{
				CheckName: name,
				Timestamp: base.Add(time.Duration(col) * time.Minute),
				Status:    s,
			})
		}
	}
	return entries
}

func TestCorrelateChecks_Empty(t *testing.T) {
	results := CorrelateChecks(nil, 2)
	if len(results) != 0 {
		t.Fatalf("expected no results, got %d", len(results))
	}
}

func TestCorrelateChecks_PerfectPositive(t *testing.T) {
	base := time.Now().Truncate(time.Minute)
	statuses := [][]check.Status{
		{check.StatusOK, check.StatusWarn, check.StatusCritical, check.StatusWarn},
		{check.StatusOK, check.StatusWarn, check.StatusCritical, check.StatusWarn},
	}
	entries := buildCorrEntries([]string{"alpha", "beta"}, statuses, base)
	results := CorrelateChecks(entries, 2)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.CheckA != "alpha" || r.CheckB != "beta" {
		t.Errorf("unexpected check names: %s, %s", r.CheckA, r.CheckB)
	}
	if r.Coefficient < 0.99 {
		t.Errorf("expected near-perfect correlation, got %.4f", r.Coefficient)
	}
	if r.SampleSize != 4 {
		t.Errorf("expected sample size 4, got %d", r.SampleSize)
	}
}

func TestCorrelateChecks_InsufficientSamples(t *testing.T) {
	base := time.Now().Truncate(time.Minute)
	statuses := [][]check.Status{
		{check.StatusOK},
		{check.StatusOK},
	}
	entries := buildCorrEntries([]string{"a", "b"}, statuses, base)
	results := CorrelateChecks(entries, 3)
	if len(results) != 0 {
		t.Fatalf("expected no results due to insufficient samples, got %d", len(results))
	}
}

func TestCorrelateChecks_NoOverlap(t *testing.T) {
	base := time.Now().Truncate(time.Minute)
	entries := []Entry{
		{CheckName: "x", Timestamp: base, Status: check.StatusOK},
		{CheckName: "y", Timestamp: base.Add(10 * time.Minute), Status: check.StatusOK},
	}
	results := CorrelateChecks(entries, 2)
	if len(results) != 0 {
		t.Fatalf("expected no results with no timestamp overlap, got %d", len(results))
	}
}

func TestCorrelationResult_String(t *testing.T) {
	r := CorrelationResult{CheckA: "svc-a", CheckB: "svc-b", Coefficient: 0.85, SampleSize: 10}
	s := r.String()
	for _, want := range []string{"svc-a", "svc-b", "0.850", "strong", "n=10"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() missing %q, got: %s", want, s)
		}
	}
}

func TestCorrelationLabel(t *testing.T) {
	cases := []struct {
		r    float64
		want string
	}{
		{0.95, "very strong"},
		{-0.95, "very strong"},
		{0.75, "strong"},
		{0.55, "moderate"},
		{0.35, "weak"},
		{0.1, "negligible"},
	}
	for _, tc := range cases {
		if got := correlationLabel(tc.r); got != tc.want {
			t.Errorf("correlationLabel(%.2f) = %q, want %q", tc.r, got, tc.want)
		}
	}
}
