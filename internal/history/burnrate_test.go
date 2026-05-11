package history

import (
	"strings"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
)

func buildBurnRateEntries(name string, statuses []check.Status, base time.Time) []StoreEntry {
	entries := make([]StoreEntry, len(statuses))
	for i, s := range statuses {
		entries[i] = StoreEntry{
			CheckName: name,
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Status:    s,
		}
	}
	return entries
}

func TestComputeBurnRate_Empty(t *testing.T) {
	result := ComputeBurnRate(nil, BurnRateOptions{SLOPercent: 99.9, WindowHours: 1})
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d reports", len(result))
	}
}

func TestComputeBurnRate_ZeroWindow(t *testing.T) {
	now := time.Now().UTC()
	entries := buildBurnRateEntries("svc", []check.Status{check.StatusOK}, now.Add(-30*time.Minute))
	result := ComputeBurnRate(entries, BurnRateOptions{SLOPercent: 99.9, WindowHours: 0})
	if len(result) != 0 {
		t.Fatalf("expected empty result for zero window, got %d", len(result))
	}
}

func TestComputeBurnRate_AllOK(t *testing.T) {
	now := time.Now().UTC()
	statuses := []check.Status{check.StatusOK, check.StatusOK, check.StatusOK, check.StatusOK}
	entries := buildBurnRateEntries("api", statuses, now.Add(-50*time.Minute))

	reports := ComputeBurnRate(entries, BurnRateOptions{SLOPercent: 99.9, WindowHours: 1})
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].BurnRate != 0 {
		t.Errorf("expected burn rate 0 for all-OK, got %.4f", reports[0].BurnRate)
	}
	if reports[0].ErrorCount != 0 {
		t.Errorf("expected 0 errors, got %d", reports[0].ErrorCount)
	}
}

func TestComputeBurnRate_HighBurnRate(t *testing.T) {
	now := time.Now().UTC()
	statuses := []check.Status{
		check.StatusCritical, check.StatusCritical, check.StatusOK, check.StatusOK,
		check.StatusCritical, check.StatusOK, check.StatusOK, check.StatusOK,
		check.StatusOK, check.StatusOK,
	}
	entries := buildBurnRateEntries("db", statuses, now.Add(-55*time.Minute))

	reports := ComputeBurnRate(entries, BurnRateOptions{SLOPercent: 99.9, WindowHours: 1})
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	r := reports[0]
	if r.TotalCount != 10 {
		t.Errorf("expected 10 total, got %d", r.TotalCount)
	}
	if r.ErrorCount != 3 {
		t.Errorf("expected 3 errors, got %d", r.ErrorCount)
	}
	// observed=0.3, sloError=0.001 → burnRate ≈ 300
	if r.BurnRate < 100 {
		t.Errorf("expected high burn rate, got %.2f", r.BurnRate)
	}
}

func TestComputeBurnRate_OldEntriesExcluded(t *testing.T) {
	now := time.Now().UTC()
	old := buildBurnRateEntries("svc", []check.Status{check.StatusCritical}, now.Add(-3*time.Hour))
	recent := buildBurnRateEntries("svc", []check.Status{check.StatusOK}, now.Add(-10*time.Minute))

	reports := ComputeBurnRate(append(old, recent...), BurnRateOptions{SLOPercent: 99.9, WindowHours: 1})
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].ErrorCount != 0 {
		t.Errorf("old critical entry should be excluded; got %d errors", reports[0].ErrorCount)
	}
}

func TestBurnRateReport_String(t *testing.T) {
	r := BurnRateReport{
		CheckName:   "payments",
		BurnRate:    4.25,
		ErrorCount:  5,
		TotalCount:  20,
		WindowStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		WindowEnd:   time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
	}
	s := r.String()
	for _, want := range []string{"payments", "4.25", "5/20"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() missing %q: %s", want, s)
		}
	}
}
