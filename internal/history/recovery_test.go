package history

import (
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
)

func buildRecoveryEntries() []Entry {
	now := time.Now().UTC().Truncate(time.Second)
	return []Entry{
		{CheckName: "api", Timestamp: now.Add(-10 * time.Minute), Status: check.StatusOK},
		{CheckName: "api", Timestamp: now.Add(-8 * time.Minute), Status: check.StatusCritical},
		{CheckName: "api", Timestamp: now.Add(-6 * time.Minute), Status: check.StatusCritical},
		{CheckName: "api", Timestamp: now.Add(-4 * time.Minute), Status: check.StatusOK},
		{CheckName: "db", Timestamp: now.Add(-9 * time.Minute), Status: check.StatusWarn},
		{CheckName: "db", Timestamp: now.Add(-7 * time.Minute), Status: check.StatusOK},
	}
}

func TestDetectRecoveries_Empty(t *testing.T) {
	rpt := DetectRecoveries(nil)
	if rpt.TotalEvents != 0 {
		t.Fatalf("expected 0 events, got %d", rpt.TotalEvents)
	}
}

func TestDetectRecoveries_AllOK(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		{CheckName: "api", Timestamp: now.Add(-2 * time.Minute), Status: check.StatusOK},
		{CheckName: "api", Timestamp: now.Add(-1 * time.Minute), Status: check.StatusOK},
	}
	rpt := DetectRecoveries(entries)
	if rpt.TotalEvents != 0 {
		t.Fatalf("expected 0 events, got %d", rpt.TotalEvents)
	}
}

func TestDetectRecoveries_SingleEvent(t *testing.T) {
	entries := buildRecoveryEntries()
	// only api check entries
	var api []Entry
	for _, e := range entries {
		if e.CheckName == "api" {
			api = append(api, e)
		}
	}
	rpt := DetectRecoveries(api)
	if rpt.TotalEvents != 1 {
		t.Fatalf("expected 1 event, got %d", rpt.TotalEvents)
	}
	if rpt.Records[0].PrevStatus != check.StatusCritical {
		t.Errorf("expected prev status Critical, got %v", rpt.Records[0].PrevStatus)
	}
	if rpt.Records[0].Downtime != 4*time.Minute {
		t.Errorf("expected 4m downtime, got %v", rpt.Records[0].Downtime)
	}
}

func TestDetectRecoveries_MultipleChecks(t *testing.T) {
	rpt := DetectRecoveries(buildRecoveryEntries())
	if rpt.TotalEvents != 2 {
		t.Fatalf("expected 2 events, got %d", rpt.TotalEvents)
	}
	if rpt.AvgDowntime == 0 {
		t.Error("expected non-zero avg downtime")
	}
}

func TestRecoveryReport_String_ContainsEvents(t *testing.T) {
	rpt := DetectRecoveries(buildRecoveryEntries())
	s := rpt.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
	if rpt.TotalEvents > 0 && len(s) < 10 {
		t.Errorf("string too short: %q", s)
	}
}
