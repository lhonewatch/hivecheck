package history

import (
	"strings"
	"testing"
	"time"

	"github.com/user/hivecheck/internal/check"
)

func buildScorecardEntries() []Entry {
	now := time.Now()
	return []Entry{
		{CheckName: "alpha", Status: check.StatusOK, Duration: 100 * time.Millisecond, Timestamp: now.Add(-5 * time.Minute)},
		{CheckName: "alpha", Status: check.StatusOK, Duration: 120 * time.Millisecond, Timestamp: now.Add(-4 * time.Minute)},
		{CheckName: "alpha", Status: check.StatusWarn, Duration: 200 * time.Millisecond, Timestamp: now.Add(-3 * time.Minute)},
		{CheckName: "beta", Status: check.StatusCritical, Duration: 5 * time.Second, Timestamp: now.Add(-6 * time.Minute)},
		{CheckName: "beta", Status: check.StatusCritical, Duration: 6 * time.Second, Timestamp: now.Add(-2 * time.Minute)},
		{CheckName: "beta", Status: check.StatusOK, Duration: 300 * time.Millisecond, Timestamp: now.Add(-1 * time.Minute)},
	}
}

func TestComputeScorecard_Empty(t *testing.T) {
	result := ComputeScorecard(nil, ScorecardOptions{})
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d entries", len(result))
	}
}

func TestComputeScorecard_Counts(t *testing.T) {
	entries := buildScorecardEntries()
	result := ComputeScorecard(entries, ScorecardOptions{})
	if len(result) != 2 {
		t.Fatalf("expected 2 scorecard entries, got %d", len(result))
	}
}

func TestComputeScorecard_AlphaHigherThanBeta(t *testing.T) {
	entries := buildScorecardEntries()
	result := ComputeScorecard(entries, ScorecardOptions{})
	// results are sorted descending by score
	if result[0].CheckName != "alpha" {
		t.Errorf("expected alpha to rank first, got %s", result[0].CheckName)
	}
	if result[0].Score <= result[1].Score {
		t.Errorf("alpha score %.2f should exceed beta score %.2f", result[0].Score, result[1].Score)
	}
}

func TestComputeScorecard_Availability(t *testing.T) {
	entries := buildScorecardEntries()
	result := ComputeScorecard(entries, ScorecardOptions{})
	var alpha ScorecardEntry
	for _, e := range result {
		if e.CheckName == "alpha" {
			alpha = e
		}
	}
	// 2 OK out of 3 = 66.67%
	if alpha.Availability < 66.0 || alpha.Availability > 67.0 {
		t.Errorf("unexpected availability %.2f", alpha.Availability)
	}
	if alpha.Samples != 3 {
		t.Errorf("expected 3 samples, got %d", alpha.Samples)
	}
}

func TestComputeScorecard_SinceFilter(t *testing.T) {
	entries := buildScorecardEntries()
	opts := ScorecardOptions{Since: time.Now().Add(-3 * time.Minute)}
	result := ComputeScorecard(entries, opts)
	for _, e := range result {
		if e.CheckName == "alpha" && e.Samples != 1 {
			t.Errorf("expected 1 alpha sample after filter, got %d", e.Samples)
		}
	}
}

func TestScorecardEntry_String(t *testing.T) {
	e := ScorecardEntry{
		CheckName:    "mycheck",
		Score:        87.5,
		Availability: 90.0,
		Reliability:  95.0,
		Perf:         75.0,
		LastStatus:   check.StatusOK,
		Samples:      20,
	}
	s := e.String()
	for _, want := range []string{"mycheck", "87.5", "90.0", "20"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() missing %q in %q", want, s)
		}
	}
}
