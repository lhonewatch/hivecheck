package history

import (
	"testing"
	"time"
)

func buildWindowEntries(now time.Time) []Entry {
	return []Entry{
		{CheckName: "api", Timestamp: now.Add(-10 * time.Minute), Status: StatusOK, Duration: 50 * time.Millisecond},
		{CheckName: "api", Timestamp: now.Add(-20 * time.Minute), Status: StatusOK, Duration: 60 * time.Millisecond},
		{CheckName: "api", Timestamp: now.Add(-40 * time.Minute), Status: StatusWarn, Duration: 200 * time.Millisecond},
		{CheckName: "api", Timestamp: now.Add(-90 * time.Minute), Status: StatusOK, Duration: 55 * time.Millisecond}, // outside 1h window
		{CheckName: "db", Timestamp: now.Add(-5 * time.Minute), Status: StatusCritical, Duration: 400 * time.Millisecond},
		{CheckName: "db", Timestamp: now.Add(-30 * time.Minute), Status: StatusOK, Duration: 80 * time.Millisecond},
	}
}

func TestComputeWindow_Empty(t *testing.T) {
	now := time.Now()
	got := ComputeWindow(nil, WindowHour, now)
	if len(got) != 0 {
		t.Fatalf("expected empty, got %d entries", len(got))
	}
}

func TestComputeWindow_HourFiltersOld(t *testing.T) {
	now := time.Now()
	entries := buildWindowEntries(now)
	stats := ComputeWindow(entries, WindowHour, now)

	index := make(map[string]WindowStats)
	for _, s := range stats {
		index[s.CheckName] = s
	}

	api, ok := index["api"]
	if !ok {
		t.Fatal("expected api stats")
	}
	// 3 entries within 1h (not the -90m one)
	if api.Total != 3 {
		t.Errorf("api total: want 3, got %d", api.Total)
	}
	if api.OKCount != 2 {
		t.Errorf("api ok: want 2, got %d", api.OKCount)
	}
	if api.WarnCount != 1 {
		t.Errorf("api warn: want 1, got %d", api.WarnCount)
	}
}

func TestComputeWindow_UptimePct(t *testing.T) {
	now := time.Now()
	entries := buildWindowEntries(now)
	stats := ComputeWindow(entries, WindowHour, now)

	for _, s := range stats {
		if s.CheckName == "db" {
			// 1 ok out of 2 = 50%
			if s.UptimePct != 50.0 {
				t.Errorf("db uptime: want 50.0, got %.1f", s.UptimePct)
			}
			if s.CritCount != 1 {
				t.Errorf("db crit: want 1, got %d", s.CritCount)
			}
		}
	}
}

func TestComputeWindow_AvgDuration(t *testing.T) {
	now := time.Now()
	entries := buildWindowEntries(now)
	stats := ComputeWindow(entries, WindowHour, now)

	for _, s := range stats {
		if s.CheckName == "api" {
			// (50+60+200)/3 = 103ms
			want := time.Duration((50+60+200)/3) * time.Millisecond
			if s.AvgDuration != want {
				t.Errorf("api avg duration: want %s, got %s", want, s.AvgDuration)
			}
		}
	}
}

func TestComputeWindow_DayIncludesAll(t *testing.T) {
	now := time.Now()
	entries := buildWindowEntries(now)
	stats := ComputeWindow(entries, WindowDay, now)

	for _, s := range stats {
		if s.CheckName == "api" && s.Total != 4 {
			t.Errorf("day window api total: want 4, got %d", s.Total)
		}
	}
}

func TestWindowStats_String(t *testing.T) {
	w := WindowStats{
		CheckName: "api",
		Window:    WindowHour,
		Total:     5,
		OKCount:   4,
		UptimePct: 80.0,
	}
	s := w.String()
	for _, want := range []string{"api", "1h", "total=5", "ok=4", "80.0%"} {
		if !contains(s, want) {
			t.Errorf("String() missing %q in %q", want, s)
		}
	}
}

func contains(s, sub string) bool {
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
