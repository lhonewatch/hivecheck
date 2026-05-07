package history

import (
	"fmt"

	"github.com/example/hivecheck/internal/check"
)

// Trend summarises how a named check's status has changed over stored entries.
type Trend struct {
	CheckName   string        `json:"check_name"`
	StatusCounts map[string]int `json:"status_counts"`
	LastStatus  check.Status  `json:"last_status"`
	Flapping    bool          `json:"flapping"`
}

// Analyse computes per-check trends from a slice of historical entries.
// A check is considered flapping when its status changed on more than half
// of the recorded transitions.
func Analyse(entries []Entry) []Trend {
	if len(entries) == 0 {
		return nil
	}

	// Collect per-check status sequences.
	sequences := map[string][]check.Status{}
	for _, e := range entries {
		for _, r := range e.Results {
			sequences[r.Name] = append(sequences[r.Name], r.Status)
		}
	}

	trends := make([]Trend, 0, len(sequences))
	for name, statuses := range sequences {
		counts := map[string]int{}
		for _, s := range statuses {
			counts[s.String()]++
		}

		changes := 0
		for i := 1; i < len(statuses); i++ {
			if statuses[i] != statuses[i-1] {
				changes++
			}
		}
		flapping := len(statuses) > 1 && changes > len(statuses)/2

		trends = append(trends, Trend{
			CheckName:    name,
			StatusCounts: counts,
			LastStatus:   statuses[len(statuses)-1],
			Flapping:     flapping,
		})
	}
	return trends
}

// FormatTrend returns a human-readable summary line for a Trend.
func FormatTrend(t Trend) string {
	flap := ""
	if t.Flapping {
		flap = " [FLAPPING]"
	}
	return fmt.Sprintf("%-30s last=%-8s counts=%v%s",
		t.CheckName, t.LastStatus.String(), t.StatusCounts, flap)
}
