package history

import (
	"fmt"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// Streak describes a consecutive run of the same status for a named check.
type Streak struct {
	CheckName string
	Status    check.Status
	Length    int           // number of consecutive entries
	Since     time.Time     // timestamp of the first entry in the streak
	Until     time.Time     // timestamp of the last entry in the streak
	Duration  time.Duration // wall-clock span of the streak
}

// String returns a human-readable summary of the streak.
func (s Streak) String() string {
	return fmt.Sprintf("%s: %s streak of %d entries over %s (since %s)",
		s.CheckName,
		s.Status,
		s.Length,
		s.Duration.Round(time.Second),
		s.Since.Format(time.RFC3339),
	)
}

// ComputeStreaks returns the current trailing streak for every check found in
// entries. Only the most-recent consecutive run of the same status is returned
// per check (i.e. the streak that is active at the end of the slice).
//
// entries must be ordered oldest-first; the function does not sort them.
func ComputeStreaks(entries []Entry) []Streak {
	if len(entries) == 0 {
		return nil
	}

	// last status and streak metadata keyed by check name
	type state struct {
		status check.Status
		count  int
		since  time.Time
		until  time.Time
	}

	states := make(map[string]*state)

	for _, e := range entries {
		st, ok := states[e.CheckName]
		if !ok || st.status != e.Status {
			states[e.CheckName] = &state{
				status: e.Status,
				count:  1,
				since:  e.Timestamp,
				until:  e.Timestamp,
			}
			continue
		}
		st.count++
		if e.Timestamp.Before(st.since) {
			st.since = e.Timestamp
		}
		if e.Timestamp.After(st.until) {
			st.until = e.Timestamp
		}
	}

	streaks := make([]Streak, 0, len(states))
	for name, st := range states {
		d := st.until.Sub(st.since)
		if d < 0 {
			d = 0
		}
		streaks = append(streaks, Streak{
			CheckName: name,
			Status:    st.status,
			Length:    st.count,
			Since:     st.since,
			Until:     st.until,
			Duration:  d,
		})
	}
	return streaks
}
