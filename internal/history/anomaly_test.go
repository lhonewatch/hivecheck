package history

import (
	"strings"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
)

func buildAnomalyEntries(statuses []check.Status) []Entry {
	base := time.Now().Add(-time.Duration(len(statuses)) * time.Minute)
	entries := make([]Entry, len(statuses))
	for i, s := range statuses {
		entries[i] = Entry{
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Status:    s,
		}
	}
	return entries
}

func TestDetectAnomalies_Empty(t *testing.T) {
	anomalies := DetectAnomalies("svc", nil)
	if len(anomalies) != 0 {
		t.Fatalf("expected no anomalies, got %d", len(anomalies))
	}
}

func TestDetectAnomalies_AllOK(t *testing.T) {
	ok := check.StatusOK
	entries := buildAnomalyEntries([]check.Status{ok, ok, ok, ok, ok})
	anomalies := DetectAnomalies("svc", entries)
	if len(anomalies) != 0 {
		t.Fatalf("expected no anomalies, got %d", len(anomalies))
	}
}

func TestDetectAnomalies_Spike(t *testing.T) {
	ok := check.StatusOK
	warn := check.StatusWarn
	entries := buildAnomalyEntries([]check.Status{ok, ok, warn, ok, ok})
	anomalies := DetectAnomalies("svc", entries)
	if len(anomalies) != 1 {
		t.Fatalf("expected 1 anomaly, got %d", len(anomalies))
	}
	if anomalies[0].Kind != AnomalySpike {
		t.Errorf("expected spike, got %s", anomalies[0].Kind)
	}
}

func TestDetectAnomalies_Degraded(t *testing.T) {
	ok := check.StatusOK
	crit := check.StatusCritical
	entries := buildAnomalyEntries([]check.Status{ok, crit, crit, crit, ok})
	anomalies := DetectAnomalies("svc", entries)

	kinds := map[AnomalyKind]bool{}
	for _, a := range anomalies {
		kinds[a.Kind] = true
	}
	if !kinds[AnomalyDegraded] {
		t.Error("expected degraded anomaly")
	}
	if !kinds[AnomalyRecovery] {
		t.Error("expected recovery anomaly")
	}
}

func TestDetectAnomalies_Recovery(t *testing.T) {
	ok := check.StatusOK
	warn := check.StatusWarn
	entries := buildAnomalyEntries([]check.Status{warn, warn, warn, ok})
	anomalies := DetectAnomalies("svc", entries)

	var found bool
	for _, a := range anomalies {
		if a.Kind == AnomalyRecovery {
			found = true
		}
	}
	if !found {
		t.Error("expected a recovery anomaly")
	}
}

func TestAnomaly_String_ContainsFields(t *testing.T) {
	a := Anomaly{
		CheckName: "my-service",
		Kind:      AnomalySpike,
		At:        time.Now(),
		Status:    check.StatusWarn,
		Message:   "isolated result",
	}
	s := a.String()
	for _, want := range []string{"my-service", "spike", "warn", "isolated result"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() missing %q in %q", want, s)
		}
	}
}
