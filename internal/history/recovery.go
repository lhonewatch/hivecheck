package history

import (
	"fmt"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// RecoveryRecord describes a single recovery event for a check.
type RecoveryRecord struct {
	CheckName    string
	FailedAt     time.Time
	RecoveredAt  time.Time
	Downtime     time.Duration
	PrevStatus   check.Status
}

// RecoveryReport holds all recovery events found in the entry window.
type RecoveryReport struct {
	Records      []RecoveryRecord
	TotalEvents  int
	AvgDowntime  time.Duration
}

// String returns a human-readable summary of the recovery report.
func (r RecoveryReport) String() string {
	if r.TotalEvents == 0 {
		return "recovery: no recovery events detected"
	}
	return fmt.Sprintf("recovery: %d event%s, avg downtime %s",
		r.TotalEvents, pluralSuffix(r.TotalEvents), r.AvgDowntime.Round(time.Second))
}

// DetectRecoveries scans entries for transitions from a non-OK status back to
// OK, recording the downtime for each such event per check.
func DetectRecoveries(entries []Entry) RecoveryReport {
	if len(entries) == 0 {
		return RecoveryReport{}
	}

	type state struct {
		failing    bool
		failedAt   time.Time
		lastStatus check.Status
	}

	states := make(map[string]*state)
	var records []RecoveryRecord

	for _, e := range entries {
		st, ok := states[e.CheckName]
		if !ok {
			st = &state{}
			states[e.CheckName] = st
		}

		if e.Status != check.StatusOK && !st.failing {
			st.failing = true
			st.failedAt = e.Timestamp
			st.lastStatus = e.Status
		} else if e.Status == check.StatusOK && st.failing {
			records = append(records, RecoveryRecord{
				CheckName:   e.CheckName,
				FailedAt:    st.failedAt,
				RecoveredAt: e.Timestamp,
				Downtime:    e.Timestamp.Sub(st.failedAt),
				PrevStatus:  st.lastStatus,
			})
			st.failing = false
		}
	}

	if len(records) == 0 {
		return RecoveryReport{}
	}

	var total time.Duration
	for _, r := range records {
		total += r.Downtime
	}

	return RecoveryReport{
		Records:     records,
		TotalEvents: len(records),
		AvgDowntime: time.Duration(int64(total) / int64(len(records))),
	}
}
