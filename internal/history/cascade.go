package history

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// CascadeOptions controls cascade failure detection behaviour.
type CascadeOptions struct {
	// WindowDuration is the rolling window within which failures are grouped.
	WindowDuration time.Duration
	// MinChecks is the minimum number of distinct failing checks to constitute
	// a cascade.
	MinChecks int
}

// DefaultCascadeOptions returns sensible defaults.
func DefaultCascadeOptions() CascadeOptions {
	return CascadeOptions{
		WindowDuration: 3 * time.Minute,
		MinChecks:      3,
	}
}

// CascadeEvent represents a detected cascade failure.
type CascadeEvent struct {
	// Start is the timestamp of the first failure in the cascade.
	Start time.Time
	// End is the timestamp of the last failure in the cascade.
	End time.Time
	// Checks lists the failing check names in order of first failure.
	Checks []string
}

// String returns a human-readable summary of the cascade event.
func (c CascadeEvent) String() string {
	return fmt.Sprintf("cascade at %s (+%s): %s",
		c.Start.Format(time.RFC3339),
		c.End.Sub(c.Start).Round(time.Second),
		strings.Join(c.Checks, " → "),
	)
}

// DetectCascades scans historical entries for cascade failure patterns.
// Entries need not be pre-sorted; the function sorts them internally.
func DetectCascades(entries []Entry, opts CascadeOptions) ([]CascadeEvent, error) {
	if opts.WindowDuration <= 0 {
		return nil, fmt.Errorf("cascade: WindowDuration must be positive")
	}
	if opts.MinChecks < 2 {
		return nil, fmt.Errorf("cascade: MinChecks must be at least 2")
	}

	// Collect only non-OK entries.
	type failPoint struct {
		at   time.Time
		name string
	}
	var fails []failPoint
	for _, e := range entries {
		if e.Status != StatusOK {
			fails = append(fails, failPoint{at: e.Timestamp, name: e.CheckName})
		}
	}
	sort.Slice(fails, func(i, j int) bool { return fails[i].at.Before(fails[j].at) })

	var events []CascadeEvent
	for i := 0; i < len(fails); i++ {
		win := []failPoint{fails[i]}
		seen := map[string]bool{fails[i].name: true}
		for j := i + 1; j < len(fails); j++ {
			if fails[j].at.Sub(fails[i].at) > opts.WindowDuration {
				break
			}
			if !seen[fails[j].name] {
				seen[fails[j].name] = true
				win = append(win, fails[j])
			}
		}
		if len(win) >= opts.MinChecks {
			names := make([]string, len(win))
			for k, fp := range win {
				names[k] = fp.name
			}
			events = append(events, CascadeEvent{
				Start:  win[0].at,
				End:    win[len(win)-1].at,
				Checks: names,
			})
			// Advance i past this window to avoid overlapping cascades.
			i += len(win) - 1
		}
	}
	return events, nil
}
