package history

import (
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
)

func buildRollupEntries() []StoreEntry {
	base := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	return []StoreEntry{
		{CheckName: "api", Timestamp: base, Status: check.StatusOK, Duration: 100 * time.Millisecond},
		{CheckName: "api", Timestamp: base.Add(10 * time.Minute), Status: check.StatusWarning, Duration: 200 * time.Millisecond},
		{CheckName: "api", Timestamp: base.Add(70 * time.Minute), Status: check.StatusCritical, Duration: 300 * time.Millisecond},
		{CheckName: "db", Timestamp: base, Status: check.StatusOK, Duration: 50 * time.Millisecond},
		{CheckName: "db", Timestamp: base.Add(25 * time.Hour), Status: check.StatusError, Duration: 400 * time.Millisecond},
	}
}

func TestRollupEntries_Daily(t *testing.T) {
	entries := buildRollupEntries()
	buckets := RollupEntries(entries, RollupDaily)

	// api: 2 entries on day 1, 1 on day 1+1h (still day 1 UTC), then 1 on day 1 for db
	// Expect 3 buckets: api/2024-06-01, api/2024-06-01 (same day for +70min), db/2024-06-01, db/2024-06-02
	if len(buckets) != 3 {
		t.Fatalf("expected 3 daily buckets, got %d", len(buckets))
	}
}

func TestRollupEntries_Hourly(t *testing.T) {
	entries := buildRollupEntries()
	buckets := RollupEntries(entries, RollupHourly)

	// api hour 10: 2 entries; api hour 11: 1 entry; db hour 10: 1; db +25h: 1
	if len(buckets) != 4 {
		t.Fatalf("expected 4 hourly buckets, got %d", len(buckets))
	}
}

func TestRollupEntries_AvgDuration(t *testing.T) {
	base := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	entries := []StoreEntry{
		{CheckName: "svc", Timestamp: base, Status: check.StatusOK, Duration: 100 * time.Millisecond},
		{CheckName: "svc", Timestamp: base.Add(5 * time.Minute), Status: check.StatusOK, Duration: 300 * time.Millisecond},
	}
	buckets := RollupEntries(entries, RollupDaily)
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	if buckets[0].AvgDuration != 200*time.Millisecond {
		t.Errorf("expected avg 200ms, got %s", buckets[0].AvgDuration)
	}
}

func TestRollupEntries_Empty(t *testing.T) {
	buckets := RollupEntries(nil, RollupDaily)
	if len(buckets) != 0 {
		t.Errorf("expected 0 buckets for empty input")
	}
}

func TestRollupBucket_WorstStatus(t *testing.T) {
	tests := []struct {
		bucket RollupBucket
		want   check.Status
	}{
		{RollupBucket{OKCount: 3}, check.StatusOK},
		{RollupBucket{OKCount: 2, WarnCount: 1}, check.StatusWarning},
		{RollupBucket{WarnCount: 1, CritCount: 1}, check.StatusCritical},
		{RollupBucket{CritCount: 1, ErrorCount: 1}, check.StatusError},
	}
	for _, tt := range tests {
		got := tt.bucket.WorstStatus()
		if got != tt.want {
			t.Errorf("WorstStatus() = %v, want %v", got, tt.want)
		}
	}
}

func TestRollupBucket_String(t *testing.T) {
	b := RollupBucket{
		CheckName:   "api",
		PeriodStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		Total:       5,
		OKCount:     4,
		WarnCount:   1,
		AvgDuration: 150 * time.Millisecond,
	}
	s := b.String()
	for _, want := range []string{"api", "total=5", "ok=4", "warn=1", "150ms"} {
		if !contains(s, want) {
			t.Errorf("String() missing %q in %q", want, s)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsHelper(s, sub))
}

func containsHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
