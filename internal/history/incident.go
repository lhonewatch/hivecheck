package history

import (
	"fmt"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// Incident represents a contiguous period during which a check was in a
// non-OK state.
type Incident struct {
	CheckName string
	StartedAt time.Time
	ResolvedAt time.Time // zero if still open
	PeakStatus check.Status
	Duration   time.Duration
}

// IsOpen reports whether the incident has not yet been resolved.
func (i Incident) IsOpen() bool { return i.ResolvedAt.IsZero() }

// String returns a human-readable summary of the incident.
func (i Incident) String() string {
	status := "open"
	if !i.IsOpen() {
		status = fmt.Sprintf("resolved after %s", i.Duration.Round(time.Second))
	}
	return fmt.Sprintf("Incident[%s] peak=%s started=%s (%s)",
		i.CheckName, i.PeakStatus, i.StartedAt.Format(time.RFC3339), status)
}

// DetectIncidents scans the provided entries (assumed sorted ascending by
// time) and returns all incident windows found.  An incident begins when
// a check transitions to a non-OK status and ends when it returns to OK.
func DetectIncidents(entries []Entry) []Incident {
	type open struct {
		start time.Time
		peak  check.Status
		last  time.Time
	}

	active := map[string]*open{}
	var incidents []Incident

	for _, e := range entries {
		name := e.CheckName
		if e.Status == check.StatusOK {
			if o, ok := active[name]; ok {
				incidents = append(incidents, Incident{
					CheckName:  name,
					StartedAt:  o.start,
					ResolvedAt: e.Timestamp,
					PeakStatus: o.peak,
					Duration:   e.Timestamp.Sub(o.start),
				})
				delete(active, name)
			}
			continue
		}
		if o, ok := active[name]; ok {
			if e.Status > o.peak {
				o.peak = e.Status
			}
			o.last = e.Timestamp
		} else {
			active[name] = &open{start: e.Timestamp, peak: e.Status, last: e.Timestamp}
		}
	}

	// Emit still-open incidents.
	for name, o := range active {
		incidents = append(incidents, Incident{
			CheckName:  name,
			StartedAt:  o.start,
			PeakStatus: o.peak,
			Duration:   o.last.Sub(o.start),
		})
	}
	return incidents
}
