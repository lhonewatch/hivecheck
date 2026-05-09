package history

import (
	"strings"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
)

func buildMTTREntries(name string, statuses []check.Status, base time.Time) []HistoryEntry {
	entries := make([]HistoryEntry, len(statuses))
	for i, s := range statuses {
		entries[i] = HistoryEntry{
			CheckName: name,
			Status:    s,
			Timestamp: base.Add(time.Duration(i) * time.Minute),
		}
	}
	return entries
}

func TestComputeMTTR_Empty(t *testing.T) {
	reports := ComputeMTTR(nil)
	if len(reports) != 0 {
		t.Fatalf("expected 0 reports, got %d", len(reports))
	}
}

func TestComputeMTTR_NoIncidents(t *testing.T) {
	base := time.Now()
	entries := buildMTTREntries("svc", []check.Status{
		check.StatusOK, check.StatusOK, check.StatusOK,
	}, base)

	reports := ComputeMTTR(entries)
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].Incidents != 0 {
		t.Errorf("expected 0 incidents, got %d", reports[0].Incidents)
	}
}

func TestComputeMTTR_SingleIncident(t *testing.T) {
	base := time.Now()
	// OK -> Critical (t+1m) -> OK (t+3m): downtime = 2m
	entries := buildMTTREntries("svc", []check.Status{
		check.StatusOK, check.StatusCritical, check.StatusCritical, check.StatusOK,
	}, base)

	reports := ComputeMTTR(entries)
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	r := reports[0]
	if r.Incidents != 1 {
		t.Errorf("expected 1 incident, got %d", r.Incidents)
	}
	if r.MTTR != 2*time.Minute {
		t.Errorf("expected MTTR=2m, got %s", r.MTTR)
	}
	if r.TotalDowntime != 2*time.Minute {
		t.Errorf("expected TotalDowntime=2m, got %s", r.TotalDowntime)
	}
}

func TestComputeMTTR_MultipleIncidents(t *testing.T) {
	base := time.Now()
	// incident1: t+1m -> t+2m (1m), incident2: t+4m -> t+6m (2m) => MTTR=1.5m
	entries := buildMTTREntries("svc", []check.Status{
		check.StatusOK,
		check.StatusError,
		check.StatusOK,
		check.StatusOK,
		check.StatusCritical,
		check.StatusCritical,
		check.StatusOK,
	}, base)

	reports := ComputeMTTR(entries)
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	r := reports[0]
	if r.Incidents != 2 {
		t.Errorf("expected 2 incidents, got %d", r.Incidents)
	}
	expectedMTTR := (1*time.Minute + 2*time.Minute) / 2
	if r.MTTR != expectedMTTR {
		t.Errorf("expected MTTR=%s, got %s", expectedMTTR, r.MTTR)
	}
}

func TestMTTRReport_String_ContainsName(t *testing.T) {
	r := MTTRReport{CheckName: "api-health", Incidents: 1, MTTR: 5 * time.Minute, TotalDowntime: 5 * time.Minute}
	if !strings.Contains(r.String(), "api-health") {
		t.Errorf("expected String() to contain check name, got: %s", r.String())
	}
}

func TestMTTRReport_String_NoIncidents(t *testing.T) {
	r := MTTRReport{CheckName: "db", Incidents: 0}
	if !strings.Contains(r.String(), "no incidents") {
		t.Errorf("expected 'no incidents' in output, got: %s", r.String())
	}
}
