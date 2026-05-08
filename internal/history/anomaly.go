package history

import (
	"fmt"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// AnomalyKind classifies the type of anomaly detected.
type AnomalyKind string

const (
	AnomalySpike    AnomalyKind = "spike"    // single outlier result surrounded by different status
	AnomalyDegraded AnomalyKind = "degraded" // sustained non-OK run
	AnomalyRecovery AnomalyKind = "recovery" // transition from non-OK back to OK
)

// Anomaly describes a detected anomaly within a check's history.
type Anomaly struct {
	CheckName string
	Kind      AnomalyKind
	At        time.Time
	Status    check.Status
	Message   string
}

// String returns a human-readable description of the anomaly.
func (a Anomaly) String() string {
	return fmt.Sprintf("[%s] %s anomaly for %q at %s: %s",
		a.Status, a.Kind, a.CheckName, a.At.Format(time.RFC3339), a.Message)
}

// DetectAnomalies scans a slice of entries (oldest-first) for a single check
// and returns any anomalies found.
//
// Rules:
//   - spike: a non-OK entry flanked on both sides by OK entries.
//   - degraded: three or more consecutive non-OK entries.
//   - recovery: a transition from non-OK to OK.
func DetectAnomalies(name string, entries []Entry) []Anomaly {
	if len(entries) == 0 {
		return nil
	}

	var anomalies []Anomaly
	consecutiveNonOK := 0

	for i, e := range entries {
		if e.Status == check.StatusOK {
			if consecutiveNonOK >= 3 {
				// recovery after sustained degradation
				anomalies = append(anomalies, Anomaly{
					CheckName: name,
					Kind:      AnomalyRecovery,
					At:        e.Timestamp,
					Status:    check.StatusOK,
					Message:   fmt.Sprintf("recovered after %d consecutive non-OK results", consecutiveNonOK),
				})
			}
			consecutiveNonOK = 0
			continue
		}

		// non-OK entry
		consecutiveNonOK++

		if consecutiveNonOK == 3 {
			anomalies = append(anomalies, Anomaly{
				CheckName: name,
				Kind:      AnomalyDegraded,
				At:        e.Timestamp,
				Status:    e.Status,
				Message:   "sustained degradation detected (3+ consecutive non-OK)",
			})
		}

		// spike: non-OK flanked by OK on both sides
		if i > 0 && i < len(entries)-1 &&
			entries[i-1].Status == check.StatusOK &&
			entries[i+1].Status == check.StatusOK {
			anomalies = append(anomalies, Anomaly{
				CheckName: name,
				Kind:      AnomalySpike,
				At:        e.Timestamp,
				Status:    e.Status,
				Message:   fmt.Sprintf("isolated %s result between two OK results", e.Status),
			})
		}
	}

	return anomalies
}
