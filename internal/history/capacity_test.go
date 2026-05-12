package history

import (
	"strings"
	"testing"
	"time"
)

func buildCapacityEntries(name string, durations []time.Duration) []Entry {
	base := time.Now().Add(-time.Duration(len(durations)) * time.Minute)
	var entries []Entry
	for i, d := range durations {
		entries = append(entries, Entry{
			CheckName: name,
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Duration:  d,
			Status:    StatusOK,
		})
	}
	return entries
}

func TestComputeCapacity_Empty(t *testing.T) {
	reports := ComputeCapacity(nil, nil)
	if len(reports) != 0 {
		t.Fatalf("expected no reports, got %d", len(reports))
	}
}

func TestComputeCapacity_InsufficientSamples(t *testing.T) {
	entries := buildCapacityEntries("svc", []time.Duration{
		100 * time.Millisecond,
		120 * time.Millisecond,
	})
	reports := ComputeCapacity(entries, nil)
	if len(reports) != 0 {
		t.Fatalf("expected 0 reports due to insufficient samples, got %d", len(reports))
	}
}

func TestComputeCapacity_StableHeadroom(t *testing.T) {
	durations := []time.Duration{
		200 * time.Millisecond,
		200 * time.Millisecond,
		200 * time.Millisecond,
		200 * time.Millisecond,
		200 * time.Millisecond,
	}
	entries := buildCapacityEntries("api", durations)
	opts := &CapacityOptions{CriticalCeiling: 1 * time.Second, MinSamples: 5}
	reports := ComputeCapacity(entries, opts)
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	r := reports[0]
	if r.CheckName != "api" {
		t.Errorf("unexpected check name: %s", r.CheckName)
	}
	if r.HeadroomPct <= 0 {
		t.Errorf("expected positive headroom, got %.2f", r.HeadroomPct)
	}
}

func TestComputeCapacity_DegradingHeadroom(t *testing.T) {
	durations := []time.Duration{
		100 * time.Millisecond,
		500 * time.Millisecond,
		1000 * time.Millisecond,
		2000 * time.Millisecond,
		3500 * time.Millisecond,
	}
	entries := buildCapacityEntries("slow-svc", durations)
	opts := &CapacityOptions{CriticalCeiling: 5 * time.Second, MinSamples: 5}
	reports := ComputeCapacity(entries, opts)
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].TrendSlope <= 0 {
		t.Errorf("expected positive slope for degrading check")
	}
}

func TestComputeCapacity_SortedByHeadroom(t *testing.T) {
	var entries []Entry
	entries = append(entries, buildCapacityEntries("fast", []time.Duration{
		50 * time.Millisecond, 50 * time.Millisecond, 50 * time.Millisecond,
		50 * time.Millisecond, 50 * time.Millisecond,
	})...)
	entries = append(entries, buildCapacityEntries("slow", []time.Duration{
		800 * time.Millisecond, 900 * time.Millisecond, 950 * time.Millisecond,
		970 * time.Millisecond, 990 * time.Millisecond,
	})...)
	opts := &CapacityOptions{CriticalCeiling: 1 * time.Second, MinSamples: 5}
	reports := ComputeCapacity(entries, opts)
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}
	if reports[0].HeadroomPct > reports[1].HeadroomPct {
		t.Errorf("reports should be sorted ascending by headroom")
	}
}

func TestCapacityReport_String_ContainsName(t *testing.T) {
	r := CapacityReport{
		CheckName:       "my-check",
		AvgDuration:     300 * time.Millisecond,
		TrendSlope:      1.5,
		Projected30:     345 * time.Millisecond,
		HeadroomPct:     93.1,
		CriticalCeiling: 5 * time.Second,
	}
	s := r.String()
	if !strings.Contains(s, "my-check") {
		t.Errorf("String() missing check name: %s", s)
	}
	if !strings.Contains(s, "93.1") {
		t.Errorf("String() missing headroom: %s", s)
	}
}
