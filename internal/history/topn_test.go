package history

import (
	"strings"
	"testing"
	"time"
)

func buildTopNEntries() []StoreEntry {
	now := time.Now()
	entries := []StoreEntry{}
	// alpha: 4 checks, 3 failures, avg 200ms
	for i := 0; i < 4; i++ {
		st := StatusCritical
		if i == 3 {
			st = StatusOK
		}
		entries = append(entries, StoreEntry{
			CheckName: "alpha",
			Status:    st,
			Duration:  200 * time.Millisecond,
			Timestamp: now.Add(-time.Duration(i) * time.Minute),
		})
	}
	// beta: 2 checks, 1 failure, avg 500ms
	for i := 0; i < 2; i++ {
		st := StatusOK
		if i == 0 {
			st = StatusWarn
		}
		entries = append(entries, StoreEntry{
			CheckName: "beta",
			Status:    st,
			Duration:  500 * time.Millisecond,
			Timestamp: now.Add(-time.Duration(i) * time.Minute),
		})
	}
	// gamma: 3 checks, 0 failures, avg 50ms
	for i := 0; i < 3; i++ {
		entries = append(entries, StoreEntry{
			CheckName: "gamma",
			Status:    StatusOK,
			Duration:  50 * time.Millisecond,
			Timestamp: now.Add(-time.Duration(i) * time.Minute),
		})
	}
	return entries
}

func TestComputeTopN_Empty(t *testing.T) {
	r := ComputeTopN(nil, 5, TopNByFailureRate)
	if len(r.Ranked) != 0 {
		t.Fatalf("expected empty, got %d", len(r.Ranked))
	}
}

func TestComputeTopN_ByFailureRate(t *testing.T) {
	entries := buildTopNEntries()
	r := ComputeTopN(entries, 3, TopNByFailureRate)
	if len(r.Ranked) != 3 {
		t.Fatalf("expected 3 ranked entries, got %d", len(r.Ranked))
	}
	// alpha has 75% failure rate — should be first
	if r.Ranked[0].CheckName != "alpha" {
		t.Errorf("expected alpha first, got %s", r.Ranked[0].CheckName)
	}
	if r.Ranked[0].FailureRate < 0.74 || r.Ranked[0].FailureRate > 0.76 {
		t.Errorf("unexpected failure rate %.2f", r.Ranked[0].FailureRate)
	}
}

func TestComputeTopN_ByAvgDuration(t *testing.T) {
	entries := buildTopNEntries()
	r := ComputeTopN(entries, 2, TopNByAvgDuration)
	if r.Ranked[0].CheckName != "beta" {
		t.Errorf("expected beta first by duration, got %s", r.Ranked[0].CheckName)
	}
}

func TestComputeTopN_NClampedToAvailable(t *testing.T) {
	entries := buildTopNEntries()
	r := ComputeTopN(entries, 100, TopNByFailureRate)
	if len(r.Ranked) != 3 {
		t.Errorf("expected 3 (all checks), got %d", len(r.Ranked))
	}
}

func TestComputeTopN_ZeroNDefaultsFive(t *testing.T) {
	entries := buildTopNEntries()
	r := ComputeTopN(entries, 0, TopNByFailureRate)
	// only 3 unique checks exist, so clamped to 3
	if len(r.Ranked) != 3 {
		t.Errorf("expected 3, got %d", len(r.Ranked))
	}
}

func TestTopNResult_Format_ContainsHeaders(t *testing.T) {
	entries := buildTopNEntries()
	r := ComputeTopN(entries, 3, TopNByFailureRate)
	out := r.Format()
	for _, want := range []string{"check", "fail%", "avg_dur", "alpha"} {
		if !strings.Contains(out, want) {
			t.Errorf("Format() missing %q\n%s", want, out)
		}
	}
}

func TestTopNResult_Format_Empty(t *testing.T) {
	r := TopNResult{}
	if !strings.Contains(r.Format(), "no data") {
		t.Error("expected 'no data' for empty result")
	}
}
