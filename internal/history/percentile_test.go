package history

import (
	"strings"
	"testing"
	"time"
)

func buildPercentileEntries() []Entry {
	base := time.Now()
	durations := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
		60 * time.Millisecond,
		70 * time.Millisecond,
		80 * time.Millisecond,
		90 * time.Millisecond,
		100 * time.Millisecond,
	}
	entries := make([]Entry, len(durations))
	for i, d := range durations {
		entries[i] = Entry{
			CheckName: "svc-a",
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Duration:  d,
			Status:    0,
		}
	}
	return entries
}

func TestComputePercentiles_Empty(t *testing.T) {
	results := ComputePercentiles(nil)
	if len(results) != 0 {
		t.Fatalf("expected empty result, got %d", len(results))
	}
}

func TestComputePercentiles_Values(t *testing.T) {
	entries := buildPercentileEntries()
	results := ComputePercentiles(entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.CheckName != "svc-a" {
		t.Errorf("unexpected check name: %s", r.CheckName)
	}
	if r.SampleCount != 10 {
		t.Errorf("expected 10 samples, got %d", r.SampleCount)
	}
	if r.P50 != 50*time.Millisecond {
		t.Errorf("expected p50=50ms, got %s", r.P50)
	}
	if r.P90 != 90*time.Millisecond {
		t.Errorf("expected p90=90ms, got %s", r.P90)
	}
	if r.P99 != 100*time.Millisecond {
		t.Errorf("expected p99=100ms, got %s", r.P99)
	}
}

func TestComputePercentiles_SkipsZeroDuration(t *testing.T) {
	entries := buildPercentileEntries()
	// Prepend a zero-duration entry that should be ignored.
	entries = append([]Entry{{CheckName: "svc-a", Duration: 0}}, entries...)
	results := ComputePercentiles(entries)
	if results[0].SampleCount != 10 {
		t.Errorf("zero-duration entry should be excluded; got %d samples", results[0].SampleCount)
	}
}

func TestComputePercentiles_MultipleChecks(t *testing.T) {
	base := buildPercentileEntries()
	for i := range base {
		base[i].CheckName = "svc-b"
	}
	all := append(buildPercentileEntries(), base...)
	results := ComputePercentiles(all)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].CheckName != "svc-a" || results[1].CheckName != "svc-b" {
		t.Error("results should be sorted by check name")
	}
}

func TestPercentileResult_String(t *testing.T) {
	r := PercentileResult{
		CheckName:   "svc-a",
		P50:         50 * time.Millisecond,
		P90:         90 * time.Millisecond,
		P95:         95 * time.Millisecond,
		P99:         99 * time.Millisecond,
		SampleCount: 10,
	}
	s := r.String()
	for _, want := range []string{"svc-a", "n=10", "p50", "p99"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() missing %q in %q", want, s)
		}
	}
}
