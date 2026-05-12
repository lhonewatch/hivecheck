package history

import (
	"strings"
	"testing"
	"time"
)

func buildVelocityEntries(name string, statuses []int, base time.Time) []StoreEntry {
	var out []StoreEntry
	for i, s := range statuses {
		out = append(out, StoreEntry{
			CheckName:  name,
			Status:     s,
			DurationMs: int64(100 + i*10),
			Timestamp:  base.Add(time.Duration(i) * 15 * time.Minute),
		})
	}
	return out
}

func TestComputeVelocity_Empty(t *testing.T) {
	results := ComputeVelocity(nil, time.Hour)
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestComputeVelocity_ZeroWindow(t *testing.T) {
	base := time.Now().Add(-30 * time.Minute)
	es := buildVelocityEntries("svc", []int{0, 0, 2, 2, 2, 2, 2, 2}, base)
	results := ComputeVelocity(es, 0)
	if len(results) != 0 {
		t.Fatalf("expected 0 results for zero window")
	}
}

func TestComputeVelocity_InsufficientSamples(t *testing.T) {
	base := time.Now().Add(-30 * time.Minute)
	es := buildVelocityEntries("svc", []int{0, 2, 2}, base)
	results := ComputeVelocity(es, 2*time.Hour)
	if len(results) != 0 {
		t.Fatalf("expected 0 results for fewer than 4 samples")
	}
}

func TestComputeVelocity_StableOK(t *testing.T) {
	base := time.Now().Add(-3 * time.Hour)
	es := buildVelocityEntries("svc", []int{0, 0, 0, 0, 0, 0, 0, 0}, base)
	results := ComputeVelocity(es, 6*time.Hour)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.FailureRatePerHour != 0 {
		t.Errorf("expected zero failure velocity, got %f", r.FailureRatePerHour)
	}
	if r.Accelerating {
		t.Errorf("stable-OK check should not be accelerating")
	}
}

func TestComputeVelocity_DegradingAccelerating(t *testing.T) {
	base := time.Now().Add(-4 * time.Hour)
	// first half mostly OK, second half all failing
	statuses := []int{0, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 2}
	es := buildVelocityEntries("api", statuses, base)
	results := ComputeVelocity(es, 8*time.Hour)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.FailureRatePerHour <= 0 {
		t.Errorf("expected positive failure velocity, got %f", r.FailureRatePerHour)
	}
}

func TestComputeVelocity_MultipleChecks(t *testing.T) {
	base := time.Now().Add(-2 * time.Hour)
	var es []StoreEntry
	es = append(es, buildVelocityEntries("alpha", []int{0, 0, 0, 0, 0, 0, 0, 0}, base)...)
	es = append(es, buildVelocityEntries("beta", []int{0, 2, 2, 2, 2, 2, 2, 2}, base)...)
	results := ComputeVelocity(es, 4*time.Hour)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestVelocityResult_String_ContainsName(t *testing.T) {
	v := VelocityResult{CheckName: "mysvc", FailureRatePerHour: 0.5, DurationDeltaMs: 12.3, Samples: 8}
	s := v.String()
	if !strings.Contains(s, "mysvc") {
		t.Errorf("String() missing check name: %s", s)
	}
	if !strings.Contains(s, "0.500") {
		t.Errorf("String() missing failure velocity: %s", s)
	}
}

func TestVelocityResult_String_Accelerating(t *testing.T) {
	v := VelocityResult{CheckName: "svc", Accelerating: true, Samples: 4}
	if !strings.Contains(v.String(), "ACCELERATING") {
		t.Errorf("expected ACCELERATING tag in string output")
	}
}
