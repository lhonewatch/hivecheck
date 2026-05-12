package history

import (
	"fmt"
	"math"
	"time"
)

// VelocityResult holds the rate-of-change metrics for a single check.
type VelocityResult struct {
	CheckName      string
	// FailureRatePerHour is the change in failure rate per hour over the window.
	FailureRatePerHour float64
	// DurationDeltaMs is the average change in duration (ms) per hour.
	DurationDeltaMs float64
	// Accelerating is true when the failure rate is increasing faster than the
	// previous half-window.
	Accelerating bool
	Samples        int
}

// String returns a human-readable summary of the velocity result.
func (v VelocityResult) String() string {
	acc := ""
	if v.Accelerating {
		acc = " [ACCELERATING]"
	}
	return fmt.Sprintf("%s: failure velocity %.3f/hr, duration delta %.1fms/hr (n=%d)%s",
		v.CheckName, v.FailureRatePerHour, v.DurationDeltaMs, v.Samples, acc)
}

// ComputeVelocity calculates the rate of change of failure rate and duration
// for each check over the supplied entries. Only entries within the given window
// are considered. A minimum of 4 samples is required per check.
func ComputeVelocity(entries []StoreEntry, window time.Duration) []VelocityResult {
	if len(entries) == 0 || window <= 0 {
		return nil
	}

	cutoff := time.Now().Add(-window)
	byCheck := make(map[string][]StoreEntry)
	for _, e := range entries {
		if e.Timestamp.After(cutoff) {
			byCheck[e.CheckName] = append(byCheck[e.CheckName], e)
		}
	}

	var results []VelocityResult
	for name, es := range byCheck {
		if len(es) < 4 {
			continue
		}
		sortEntriesByTime(es)

		failVel, durVel := rateOfChange(es)
		mid := len(es) / 2
		firstHalfFail, _ := rateOfChange(es[:mid])
		secondHalfFail, _ := rateOfChange(es[mid:])
		acc := secondHalfFail > firstHalfFail && secondHalfFail > 0

		results = append(results, VelocityResult{
			CheckName:          name,
			FailureRatePerHour: math.Round(failVel*1000) / 1000,
			DurationDeltaMs:    math.Round(durVel*10) / 10,
			Accelerating:       acc,
			Samples:            len(es),
		})
	}
	return results
}

// rateOfChange returns failure-rate/hr and duration-delta ms/hr for a slice.
func rateOfChange(es []StoreEntry) (failPerHr, durPerHr float64) {
	if len(es) < 2 {
		return 0, 0
	}
	span := es[len(es)-1].Timestamp.Sub(es[0].Timestamp).Hours()
	if span == 0 {
		return 0, 0
	}
	firstFail := failRate(es[:len(es)/2])
	lastFail := failRate(es[len(es)/2:])
	failPerHr = (lastFail - firstFail) / span

	firstDur := avgDur(es[:len(es)/2])
	lastDur := avgDur(es[len(es)/2:])
	durPerHr = (lastDur - firstDur) / span
	return
}

func failRate(es []StoreEntry) float64 {
	if len(es) == 0 {
		return 0
	}
	var bad float64
	for _, e := range es {
		if e.Status > 1 {
			bad++
		}
	}
	return bad / float64(len(es))
}

func avgDur(es []StoreEntry) float64 {
	if len(es) == 0 {
		return 0
	}
	var sum float64
	for _, e := range es {
		sum += float64(e.DurationMs)
	}
	return sum / float64(len(es))
}
