package history

import (
	"strings"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
)

func buildIncidentEntries(name string, statuses []check.Status, base time.Time) []Entry {
	out := make([]Entry, len(statuses))
	for i, s := range statuses {
		out[i] = Entry{
			CheckName: name,
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Status:    s,
		}
	}
	return out
}

func TestDetectIncidents_Empty(t *testing.T) {
	if got := DetectIncidents(nil); len(got) != 0 {
		t.Fatalf("expected 0 incidents, got %d", len(got))
	}
}

func TestDetectIncidents_AllOK(t *testing.T) {
	base := time.Now()
	entries := buildIncidentEntries("svc", []check.Status{
		check.StatusOK, check.StatusOK, check.StatusOK,
	}, base)
	if got := DetectIncidents(entries); len(got) != 0 {
		t.Fatalf("expected 0 incidents, got %d", len(got))
	}
}

func TestDetectIncidents_SingleResolved(t *testing.T) {
	base := time.Now()
	entries := buildIncidentEntries("svc", []check.Status{
		check.StatusOK, check.StatusWarn, check.StatusCritical, check.StatusOK,
	}, base)
	got := DetectIncidents(entries)
	if len(got) != 1 {
		t.Fatalf("expected 1 incident, got %d", len(got))
	}
	inc := got[0]
	if inc.IsOpen() {
		t.Error("expected incident to be resolved")
	}
	if inc.PeakStatus != check.StatusCritical {
		t.Errorf("expected peak Critical, got %s", inc.PeakStatus)
	}
	if inc.Duration != 2*time.Minute {
		t.Errorf("unexpected duration %s", inc.Duration)
	}
}

func TestDetectIncidents_OpenIncident(t *testing.T) {
	base := time.Now()
	entries := buildIncidentEntries("svc", []check.Status{
		check.StatusOK, check.StatusWarn, check.StatusWarn,
	}, base)
	got := DetectIncidents(entries)
	if len(got) != 1 {
		t.Fatalf("expected 1 incident, got %d", len(got))
	}
	if !got[0].IsOpen() {
		t.Error("expected incident to still be open")
	}
}

func TestIncident_String_ContainsFields(t *testing.T) {
	inc := Incident{
		CheckName:  "api",
		StartedAt:  time.Now(),
		PeakStatus: check.StatusCritical,
	}
	s := inc.String()
	for _, want := range []string{"api", "critical", "open"} {
		if !strings.Contains(strings.ToLower(s), want) {
			t.Errorf("String() missing %q: %s", want, s)
		}
	}
}
