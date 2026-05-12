package history

import (
	"math"
	"testing"
	"time"
)

func buildSpikeEntries(name string, durationsMs []int64) []Entry {
	base := time.Now().Add(-time.Duration(len(durationsMs)) * time.Minute)
	entries := make([]Entry, len(durationsMs))
	for i, d := range durationsMs {
		entries[i] = Entry{
			CheckName:  name,
			Timestamp:  base.Add(time.Duration(i) * time.Minute),
			DurationMs: d,
			Status:     0,
		}
	}
	return entries
}

func TestDetectSpikes_Empty(t *testing.T) {
	results := DetectSpikes(nil, DefaultSpikeOptions())
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestDetectSpikes_InsufficientSamples(t *testing.T) {
	entries := buildSpikeEntries("svc", []int64{10, 12, 11})
	opts := DefaultSpikeOptions()
	opts.MinSamples = 5
	results := DetectSpikes(entries, opts)
	if len(results) != 0 {
		t.Fatalf("expected 0 results for insufficient samples, got %d", len(results))
	}
}

func TestDetectSpikes_NoSpike(t *testing.T) {
	// Stable latencies — last value is within normal range.
	entries := buildSpikeEntries("svc", []int64{100, 102, 98, 101, 99, 100})
	opts := DefaultSpikeOptions()
	opts.MinSamples = 5
	results := DetectSpikes(entries, opts)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].IsSpike {
		t.Errorf("expected no spike, got one (z=%.2f)", results[0].ZScore)
	}
}

func TestDetectSpikes_SpikeDetected(t *testing.T) {
	// Last value is a large outlier.
	entries := buildSpikeEntries("svc", []int64{100, 102, 98, 101, 99, 950})
	opts := DefaultSpikeOptions()
	opts.MinSamples = 5
	opts.ZScoreThreshold = 2.5
	results := DetectSpikes(entries, opts)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].IsSpike {
		t.Errorf("expected spike to be detected (z=%.2f)", results[0].ZScore)
	}
	if results[0].LatestMs != 950 {
		t.Errorf("expected LatestMs=950, got %.1f", results[0].LatestMs)
	}
}

func TestDetectSpikes_MultipleChecks(t *testing.T) {
	a := buildSpikeEntries("alpha", []int64{100, 100, 100, 100, 100, 100})
	b := buildSpikeEntries("beta", []int64{50, 52, 48, 51, 49, 800})
	results := DetectSpikes(append(a, b...), DefaultSpikeOptions())
	spikes := 0
	for _, r := range results {
		if r.IsSpike {
			spikes++
		}
	}
	if spikes != 1 {
		t.Errorf("expected 1 spike across checks, got %d", spikes)
	}
}

func TestSpikeResult_String_NoSpike(t *testing.T) {
	r := SpikeResult{CheckName: "svc", IsSpike: false, ZScore: 0.5, MeanMs: 100}
	s := r.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
	if contains := "no spike"; !stringContains(s, contains) {
		t.Errorf("expected %q in %q", contains, s)
	}
}

func TestSpikeResult_String_Spike(t *testing.T) {
	r := SpikeResult{CheckName: "svc", IsSpike: true, ZScore: 3.1, MeanMs: 100, StdDevMs: 5, LatestMs: 900}
	s := r.String()
	if !stringContains(s, "SPIKE") {
		t.Errorf("expected SPIKE in %q", s)
	}
}

func TestDefaultSpikeOptions(t *testing.T) {
	opts := DefaultSpikeOptions()
	if opts.MinSamples <= 0 {
		t.Error("MinSamples should be positive")
	}
	if opts.ZScoreThreshold <= 0 {
		t.Error("ZScoreThreshold should be positive")
	}
	if opts.Window <= 0 {
		t.Error("Window should be positive")
	}
}

// stringContains is a small helper to avoid importing strings in tests.
func stringContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

// mathSqrt is referenced from spike.go; ensure it is available in tests.
var _ = math.Sqrt
