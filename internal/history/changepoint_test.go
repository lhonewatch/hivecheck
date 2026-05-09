package history

import (
	"strings"
	"testing"
	"time"

	"github.com/your-org/hivecheck/internal/check"
)

func buildChangepointEntries(name string, durations []time.Duration) []Entry {
	base := time.Now().Add(-time.Duration(len(durations)) * time.Minute)
	entries := make([]Entry, len(durations))
	for i, d := range durations {
		entries[i] = Entry{
			CheckName: name,
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Status:    check.StatusOK,
			Duration:  d,
		}
	}
	return entries
}

func ms(n int) time.Duration { return time.Duration(n) * time.Millisecond }

func TestDetectChangepoints_Empty(t *testing.T) {
	result := DetectChangepoints(nil, 10)
	if len(result) != 0 {
		t.Fatalf("expected no changepoints, got %d", len(result))
	}
}

func TestDetectChangepoints_InsufficientSamples(t *testing.T) {
	entries := buildChangepointEntries("svc", []time.Duration{ms(10), ms(12), ms(11)})
	result := DetectChangepoints(entries, 5)
	if len(result) != 0 {
		t.Fatalf("expected no changepoints for < 4 samples, got %d", len(result))
	}
}

func TestDetectChangepoints_NoDegradation(t *testing.T) {
	// Stable durations — delta well below threshold.
	durations := []time.Duration{ms(50), ms(51), ms(50), ms(52), ms(49), ms(51)}
	entries := buildChangepointEntries("api", durations)
	result := DetectChangepoints(entries, 20)
	if len(result) != 0 {
		t.Fatalf("expected no changepoints, got %d", len(result))
	}
}

func TestDetectChangepoints_ClearShift(t *testing.T) {
	// First half fast, second half slow — clear changepoint.
	durations := []time.Duration{
		ms(30), ms(32), ms(31), ms(30),
		ms(120), ms(125), ms(122), ms(118),
	}
	entries := buildChangepointEntries("db", durations)
	result := DetectChangepoints(entries, 50)
	if len(result) != 1 {
		t.Fatalf("expected 1 changepoint, got %d", len(result))
	}
	cp := result[0]
	if cp.CheckName != "db" {
		t.Errorf("expected check name 'db', got %q", cp.CheckName)
	}
	if cp.Delta <= 50 {
		t.Errorf("expected delta > 50ms, got %.1f", cp.Delta)
	}
}

func TestDetectChangepoints_MultipleChecks(t *testing.T) {
	fast := buildChangepointEntries("fast-svc", []time.Duration{ms(10), ms(11), ms(10), ms(11)})
	slow := []time.Duration{ms(20), ms(22), ms(200), ms(210), ms(205), ms(198)}
	degraded := buildChangepointEntries("slow-svc", slow)

	all := append(fast, degraded...)
	result := DetectChangepoints(all, 100)
	if len(result) != 1 {
		t.Fatalf("expected 1 changepoint (only slow-svc), got %d", len(result))
	}
	if result[0].CheckName != "slow-svc" {
		t.Errorf("expected slow-svc, got %q", result[0].CheckName)
	}
}

func TestChangepoint_String_ContainsFields(t *testing.T) {
	cp := Changepoint{
		CheckName: "payments",
		At:        time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		Before:    40.0,
		After:     140.0,
		Delta:     100.0,
	}
	s := cp.String()
	for _, want := range []string{"payments", "degradation", "40.0", "140.0"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() missing %q: %s", want, s)
		}
	}
}
