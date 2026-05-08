package history

import (
	"strings"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
)

func buildSLAEntries(now time.Time) []Entry {
	return []Entry{
		{CheckName: "api", Timestamp: now.Add(-1 * time.Hour), Status: check.StatusOK},
		{CheckName: "api", Timestamp: now.Add(-2 * time.Hour), Status: check.StatusOK},
		{CheckName: "api", Timestamp: now.Add(-3 * time.Hour), Status: check.StatusWarn},
		{CheckName: "api", Timestamp: now.Add(-4 * time.Hour), Status: check.StatusCritical},
		{CheckName: "db", Timestamp: now.Add(-30 * time.Minute), Status: check.StatusOK},
		{CheckName: "db", Timestamp: now.Add(-90 * time.Minute), Status: check.StatusOK},
		// outside window — should be ignored
		{CheckName: "api", Timestamp: now.Add(-48 * time.Hour), Status: check.StatusCritical},
	}
}

func TestComputeSLA_Empty(t *testing.T) {
	reports := ComputeSLA(nil, 24*time.Hour, 99.9)
	if len(reports) != 0 {
		t.Fatalf("expected 0 reports, got %d", len(reports))
	}
}

func TestComputeSLA_Counts(t *testing.T) {
	now := time.Now().UTC()
	entries := buildSLAEntries(now)
	reports := ComputeSLA(entries, 24*time.Hour, 99.0)

	byName := make(map[string]SLAReport)
	for _, r := range reports {
		byName[r.CheckName] = r
	}

	api, ok := byName["api"]
	if !ok {
		t.Fatal("missing report for 'api'")
	}
	if api.Total != 4 {
		t.Errorf("api total: want 4, got %d", api.Total)
	}
	if api.OKCount != 2 {
		t.Errorf("api ok: want 2, got %d", api.OKCount)
	}
	if api.WarnCount != 1 {
		t.Errorf("api warn: want 1, got %d", api.WarnCount)
	}
	if api.CritCount != 1 {
		t.Errorf("api crit: want 1, got %d", api.CritCount)
	}

	db := byName["db"]
	if db.Total != 2 {
		t.Errorf("db total: want 2, got %d", db.Total)
	}
}

func TestComputeSLA_Availability(t *testing.T) {
	now := time.Now().UTC()
	entries := buildSLAEntries(now)
	reports := ComputeSLA(entries, 24*time.Hour, 99.0)

	for _, r := range reports {
		if r.CheckName == "db" {
			if r.Availability != 100.0 {
				t.Errorf("db availability: want 100.0, got %.2f", r.Availability)
			}
			if !r.Compliant {
				t.Error("db should be compliant")
			}
		}
		if r.CheckName == "api" {
			if r.Compliant {
				t.Error("api should NOT be compliant at 99% target")
			}
		}
	}
}

func TestSLAReport_String_ContainsName(t *testing.T) {
	r := SLAReport{
		CheckName:    "payments",
		Window:       24 * time.Hour,
		Availability: 99.5,
		Target:       99.9,
		Compliant:    false,
	}
	s := r.String()
	for _, want := range []string{"payments", "99.50", "99.90", "NON-COMPLIANT"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() missing %q: %s", want, s)
		}
	}
}

func TestSLAReport_String_Compliant(t *testing.T) {
	r := SLAReport{
		CheckName:    "cache",
		Window:       time.Hour,
		Availability: 100.0,
		Target:       99.0,
		Compliant:    true,
	}
	if !strings.Contains(r.String(), "COMPLIANT") {
		t.Errorf("expected COMPLIANT in string: %s", r.String())
	}
}
