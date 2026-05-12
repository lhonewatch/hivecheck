package history

import (
	"strings"
	"testing"
	"time"
)

func buildPatternEntries() []Entry {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	var entries []Entry
	// 30 days of hourly entries; make hour 14 consistently degraded.
	for day := 0; day < 30; day++ {
		for hour := 0; hour < 24; hour++ {
			status := StatusOK
			if hour == 14 {
				status = StatusWarn
			}
			entries = append(entries, Entry{
				CheckName: "api",
				Timestamp: base.Add(time.Duration(day)*24*time.Hour + time.Duration(hour)*time.Hour),
				Status:    status,
			})
		}
	}
	return entries
}

func TestDetectPatterns_Empty(t *testing.T) {
	results := DetectPatterns(nil, DefaultPatternOptions())
	if len(results) != 0 {
		t.Fatalf("expected empty, got %d", len(results))
	}
}

func TestDetectPatterns_InsufficientSamples(t *testing.T) {
	entries := []Entry{
		{CheckName: "svc", Timestamp: time.Now(), Status: StatusWarn},
	}
	opts := DefaultPatternOptions()
	results := DetectPatterns(entries, opts)
	if len(results) != 0 {
		t.Fatalf("expected 0 results for insufficient samples, got %d", len(results))
	}
}

func TestDetectPatterns_HourOfDay(t *testing.T) {
	entries := buildPatternEntries()
	opts := DefaultPatternOptions()
	opts.MinConfidence = 0.1
	results := DetectPatterns(entries, opts)

	var found bool
	for _, r := range results {
		if r.Kind == PatternHourOfDay && r.Bucket == 14 && r.CheckName == "api" {
			found = true
			if r.FailRate <= 0 {
				t.Errorf("expected positive fail-rate, got %.4f", r.FailRate)
			}
			if r.Confidence <= 0 || r.Confidence > 1 {
				t.Errorf("confidence out of range: %.4f", r.Confidence)
			}
		}
	}
	if !found {
		t.Error("expected HourOfDay pattern at bucket 14 for 'api'")
	}
}

func TestPatternResult_String_ContainsFields(t *testing.T) {
	p := PatternResult{
		CheckName:  "db",
		Kind:       PatternHourOfDay,
		Bucket:     3,
		FailRate:   0.75,
		Confidence: 0.9,
	}
	s := p.String()
	for _, want := range []string{"db", "hour_of_day", "03:00", "0.75", "0.90"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() missing %q in %q", want, s)
		}
	}
}

func TestDefaultPatternOptions(t *testing.T) {
	opts := DefaultPatternOptions()
	if opts.MinConfidence <= 0 {
		t.Error("expected positive MinConfidence")
	}
	if opts.MinSamples <= 0 {
		t.Error("expected positive MinSamples")
	}
}
