package history

import (
	"fmt"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// MTTRReport holds mean-time-to-recovery statistics for a single check.
type MTTRReport struct {
	CheckName     string
	Incidents     int
	TotalDowntime time.Duration
	MTTR          time.Duration // average recovery time per incident
}

// String returns a human-readable summary of the MTTR report.
func (r MTTRReport) String() string {
	if r.Incidents == 0 {
		return fmt.Sprintf("%s: no incidents recorded", r.CheckName)
	}
	return fmt.Sprintf("%s: incidents=%d total_downtime=%s mttr=%s",
		r.CheckName, r.Incidents,
		r.TotalDowntime.Round(time.Second),
		r.MTTR.Round(time.Second),
	)
}

// ComputeMTTR calculates mean-time-to-recovery for each check found in
// entries. An "incident" begins when a check transitions to Critical or
// Error and ends when it next returns to OK. Entries must be sorted
// ascending by timestamp for accurate results.
func ComputeMTTR(entries []HistoryEntry) []MTTRReport {
	type state struct {
		inIncident bool
		incidentStart time.Time
		totalDowntime time.Duration
		incidents     int
	}

	states := make(map[string]*state)

	for _, e := range entries {
		s, ok := states[e.CheckName]
		if !ok {
			s = &state{}
			states[e.CheckName] = s
		}

		switch e.Status {
		case check.StatusCritical, check.StatusError:
			if !s.inIncident {
				s.inIncident = true
				s.incidentStart = e.Timestamp
			}
		case check.StatusOK:
			if s.inIncident {
				s.totalDowntime += e.Timestamp.Sub(s.incidentStart)
				s.incidents++
				s.inIncident = false
			}
		}
	}

	reports := make([]MTTRReport, 0, len(states))
	for name, s := range states {
		var mttr time.Duration
		if s.incidents > 0 {
			mttr = s.totalDowntime / time.Duration(s.incidents)
		}
		reports = append(reports, MTTRReport{
			CheckName:     name,
			Incidents:     s.incidents,
			TotalDowntime: s.totalDowntime,
			MTTR:          mttr,
		})
	}
	return reports
}
