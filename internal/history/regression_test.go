package history

import (
	"testing"
	"time"
)

func buildRegressionEntries(now time.Time) []Entry {
	entries := []Entry{}
	// Reference window: 25h–1h ago — mostly OK (1 fail in 10)
	base := now.Add(-25 * time.Hour)
	for i := 0; i < 10; i++ {
		status := StatusOK
		if i == 0 {
			status = StatusCritical
		}
		entries = append(entries, Entry{
			CheckName:  "api",
			Timestamp:  base.Add(time.Duration(i) * 30 * time.Minute),
			Status:     status,
			DurationMs: 100,
		})
	}
	// Recent window: last 1h — heavy failures (8 in 10) + high latency
	recent := now.Add(-50 * time.Minute)
	for i := 0; i < 10; i++ {
		status := StatusCritical
		if i >= 8 {
			status = StatusOK
		}
		entries = append(entries, Entry{
			CheckName:  "api",
			Timestamp:  recent.Add(time.Duration(i) * 5 * time.Minute),
			Status:     status,
			DurationMs: 500,
		})
	}
	return entries
}

func TestDetectRegressions_Empty(t *testing.T) {
	results := DetectRegressions(nil, DefaultRegressionOptions())
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestDetectRegressions_MajorRegression(t *testing.T) {
	now := time.Now()
	entries := buildRegressionEntries(now)
	opts := DefaultRegressionOptions()
	results := DetectRegressions(entries, opts)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.CheckName != "api" {
		t.Errorf("unexpected check name: %s", r.CheckName)
	}
	if !r.RateRegressed {
		t.Error("expected RateRegressed=true")
	}
	if r.Severity != "major" {
		t.Errorf("expected severity=major, got %s", r.Severity)
	}
	if r.CurrentRate <= r.BaselineRate {
		t.Errorf("current rate %.2f should exceed baseline %.2f", r.CurrentRate, r.BaselineRate)
	}
}

func TestDetectRegressions_NoRegression(t *testing.T) {
	now := time.Now()
	entries := []Entry{}
	// Both windows: all OK, same latency
	for i := 0; i < 20; i++ {
		entries = append(entries, Entry{
			CheckName:  "db",
			Timestamp:  now.Add(-time.Duration(25-i) * time.Hour),
			Status:     StatusOK,
			DurationMs: 50,
		})
	}
	entries = append(entries, Entry{
		CheckName:  "db",
		Timestamp:  now.Add(-10 * time.Minute),
		Status:     StatusOK,
		DurationMs: 50,
	})

	opts := DefaultRegressionOptions()
	results := DetectRegressions(entries, opts)
	for _, r := range results {
		if r.CheckName == "db" && r.Severity != "none" {
			t.Errorf("expected no regression for db, got severity=%s", r.Severity)
		}
	}
}

func TestRegressionResult_String(t *testing.T) {
	r := RegressionResult{
		CheckName:    "svc",
		Severity:     "minor",
		BaselineRate: 0.05,
		CurrentRate:  0.20,
		BaselineAvgMs: 80.0,
		CurrentAvgMs:  120.5,
	}
	s := r.String()
	if s == "" {
		t.Fatal("String() returned empty")
	}
	for _, want := range []string{"svc", "minor", "5.00", "20.00", "80.0", "120.5"} {
		if !containsStr(s, want) {
			t.Errorf("String() missing %q: %s", want, s)
		}
	}
}

func TestDefaultRegressionOptions(t *testing.T) {
	opts := DefaultRegressionOptions()
	if opts.ReferenceWindow <= 0 {
		t.Error("ReferenceWindow should be positive")
	}
	if opts.RecentWindow <= 0 {
		t.Error("RecentWindow should be positive")
	}
	if opts.RateThreshold <= 0 {
		t.Error("RateThreshold should be positive")
	}
	if opts.LatencyThresholdPct <= 0 {
		t.Error("LatencyThresholdPct should be positive")
	}
}

// containsStr is a local helper (avoids import of strings in test file).
func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()
}
