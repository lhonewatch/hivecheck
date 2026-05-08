package history

import (
	"strings"
	"testing"
	"time"
)

func buildForecastEntries(name string, count int, durationBase, durationStep float64) []Entry {
	now := time.Now()
	entries := make([]Entry, count)
	for i := 0; i < count; i++ {
		entries[i] = Entry{
			CheckName:  name,
			Timestamp:  now.Add(time.Duration(i) * time.Minute),
			DurationMS: int64(durationBase + float64(i)*durationStep),
		}
	}
	return entries
}

func TestForecastChecks_Empty(t *testing.T) {
	results := ForecastChecks(nil, time.Hour)
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestForecastChecks_ZeroHorizon(t *testing.T) {
	entries := buildForecastEntries("svc", 5, 100, 2)
	results := ForecastChecks(entries, 0)
	if len(results) != 0 {
		t.Fatalf("expected 0 results for zero horizon, got %d", len(results))
	}
}

func TestForecastChecks_InsufficientSamples(t *testing.T) {
	entries := buildForecastEntries("svc", 2, 50, 5)
	results := ForecastChecks(entries, time.Hour)
	if len(results) != 0 {
		t.Fatalf("expected 0 results for < 3 samples, got %d", len(results))
	}
}

func TestForecastChecks_Degrading(t *testing.T) {
	// Positive step → latency increasing → Degrading should be true.
	entries := buildForecastEntries("api", 10, 100, 10)
	results := ForecastChecks(entries, time.Hour)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if !r.Degrading {
		t.Error("expected Degrading=true for increasing latency")
	}
	if r.Predicted <= 100 {
		t.Errorf("expected predicted > 100, got %.2f", r.Predicted)
	}
	if r.CheckName != "api" {
		t.Errorf("unexpected check name: %s", r.CheckName)
	}
}

func TestForecastChecks_Stable(t *testing.T) {
	// Zero step → flat latency → Degrading should be false.
	entries := buildForecastEntries("db", 8, 200, 0)
	results := ForecastChecks(entries, time.Hour)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Degrading {
		t.Error("expected Degrading=false for flat latency")
	}
}

func TestForecastChecks_ConfidenceLabels(t *testing.T) {
	cases := []struct {
		count    int
		wantConf string
	}{
		{3, "low"},
		{7, "medium"},
		{20, "high"},
	}
	for _, tc := range cases {
		entries := buildForecastEntries("svc", tc.count, 50, 1)
		results := ForecastChecks(entries, time.Hour)
		if len(results) == 0 {
			t.Fatalf("count=%d: expected result", tc.count)
		}
		if results[0].Confidence != tc.wantConf {
			t.Errorf("count=%d: want confidence %q, got %q", tc.count, tc.wantConf, results[0].Confidence)
		}
	}
}

func TestForecastResult_String(t *testing.T) {
	entries := buildForecastEntries("cache", 5, 30, 3)
	results := ForecastChecks(entries, 30*time.Minute)
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	s := results[0].String()
	for _, want := range []string{"check=cache", "predicted=", "slope=", "confidence="} {
		if !strings.Contains(s, want) {
			t.Errorf("String() missing %q: %s", want, s)
		}
	}
}
