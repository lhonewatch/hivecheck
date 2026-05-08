package history

import (
	"strings"
	"testing"
	"time"
)

func buildDigestEntries(now time.Time) []Entry {
	return []Entry{
		{CheckName: "api", Status: "ok", DurationMs: 100, Timestamp: now.Add(-1 * time.Hour)},
		{CheckName: "api", Status: "warn", DurationMs: 300, Timestamp: now.Add(-2 * time.Hour)},
		{CheckName: "api", Status: "critical", DurationMs: 500, Timestamp: now.Add(-25 * time.Hour)}, // outside daily
		{CheckName: "db", Status: "ok", DurationMs: 50, Timestamp: now.Add(-3 * time.Hour)},
		{CheckName: "db", Status: "error", DurationMs: 0, Timestamp: now.Add(-4 * time.Hour)},
		{CheckName: "db", Status: "ok", DurationMs: 60, Timestamp: now.Add(-200 * time.Hour)}, // outside weekly
	}
}

func TestBuildDigest_Daily(t *testing.T) {
	now := time.Now()
	entries := buildDigestEntries(now)
	d := BuildDigest(entries, DigestDaily, now)

	if d.Period != DigestDaily {
		t.Fatalf("expected DigestDaily, got %v", d.Period)
	}
	if len(d.Entries) != 2 {
		t.Fatalf("expected 2 check entries, got %d", len(d.Entries))
	}

	api := d.Entries[0]
	if api.CheckName != "api" {
		t.Fatalf("expected api first, got %s", api.CheckName)
	}
	if api.TotalRuns != 2 {
		t.Errorf("api: expected 2 runs, got %d", api.TotalRuns)
	}
	if api.OKCount != 1 || api.WarnCount != 1 {
		t.Errorf("api: unexpected counts ok=%d warn=%d", api.OKCount, api.WarnCount)
	}
	if api.WorstStatus != "warn" {
		t.Errorf("api: expected worst=warn, got %s", api.WorstStatus)
	}
}

func TestBuildDigest_Weekly(t *testing.T) {
	now := time.Now()
	entries := buildDigestEntries(now)
	d := BuildDigest(entries, DigestWeekly, now)

	// api critical is 25h ago — inside weekly window
	var api DigestEntry
	for _, e := range d.Entries {
		if e.CheckName == "api" {
			api = e
		}
	}
	if api.TotalRuns != 3 {
		t.Errorf("weekly api: expected 3 runs, got %d", api.TotalRuns)
	}
	if api.WorstStatus != "critical" {
		t.Errorf("weekly api: expected worst=critical, got %s", api.WorstStatus)
	}
}

func TestBuildDigest_Empty(t *testing.T) {
	d := BuildDigest(nil, DigestDaily, time.Now())
	if len(d.Entries) != 0 {
		t.Errorf("expected empty digest, got %d entries", len(d.Entries))
	}
}

func TestBuildDigest_AvgDuration(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		{CheckName: "svc", Status: "ok", DurationMs: 200, Timestamp: now.Add(-1 * time.Hour)},
		{CheckName: "svc", Status: "ok", DurationMs: 400, Timestamp: now.Add(-2 * time.Hour)},
	}
	d := BuildDigest(entries, DigestDaily, now)
	if len(d.Entries) != 1 {
		t.Fatalf("expected 1 entry")
	}
	if d.Entries[0].AvgDurationMs != 300.0 {
		t.Errorf("expected avg 300, got %.1f", d.Entries[0].AvgDurationMs)
	}
}

func TestDigest_Format_ContainsHeaders(t *testing.T) {
	now := time.Now()
	entries := buildDigestEntries(now)
	d := BuildDigest(entries, DigestDaily, now)
	out := d.Format()

	for _, want := range []string{"Daily Digest", "Check", "Runs", "Worst", "api", "db"} {
		if !strings.Contains(out, want) {
			t.Errorf("Format() missing %q in output:\n%s", want, out)
		}
	}
}

func TestDigest_Format_Weekly_Label(t *testing.T) {
	d := BuildDigest(nil, DigestWeekly, time.Now())
	if !strings.Contains(d.Format(), "Weekly Digest") {
		t.Error("expected 'Weekly Digest' label in format output")
	}
}
