package history

import (
	"fmt"
	"math"
	"time"
)

// Changepoint represents a detected shift in check behaviour at a specific time.
type Changepoint struct {
	CheckName string
	At        time.Time
	Before    float64 // mean duration (ms) before the shift
	After     float64 // mean duration (ms) after the shift
	Delta     float64 // After - Before
}

// String returns a human-readable description of the changepoint.
func (c Changepoint) String() string {
	direction := "improvement"
	if c.Delta > 0 {
		direction = "degradation"
	}
	return fmt.Sprintf("%s: %s at %s (%.1fms → %.1fms, Δ%+.1fms %s)",
		c.CheckName, direction, c.At.Format(time.RFC3339),
		c.Before, c.After, c.Delta, direction)
}

// DetectChangepoints scans history entries for each check and returns any
// points where the mean response duration shifted significantly. It uses a
// simple CUSUM-style split: for each check it tries every possible split
// index and picks the one that maximises the absolute difference in means,
// reporting it only when the delta exceeds minDeltaMs.
func DetectChangepoints(entries []Entry, minDeltaMs float64) []Changepoint {
	if len(entries) == 0 {
		return nil
	}

	// Group entries by check name, preserving time order.
	groups := make(map[string][]Entry)
	for _, e := range entries {
		groups[e.CheckName] = append(groups[e.CheckName], e)
	}

	var results []Changepoint
	for name, es := range groups {
		if len(es) < 4 {
			continue
		}
		cp, ok := bestSplit(name, es, minDeltaMs)
		if ok {
			results = append(results, cp)
		}
	}
	return results
}

// bestSplit finds the split index that maximises |meanAfter - meanBefore|
// and returns a Changepoint if the delta exceeds minDeltaMs.
func bestSplit(name string, es []Entry, minDeltaMs float64) (Changepoint, bool) {
	durations := make([]float64, len(es))
	for i, e := range es {
		durations[i] = float64(e.Duration.Milliseconds())
	}

	bestDelta := 0.0
	bestIdx := -1

	for split := 1; split < len(durations)-1; split++ {
		before := mean(durations[:split])
		after := mean(durations[split:])
		if d := math.Abs(after - before); d > bestDelta {
			bestDelta = d
			bestIdx = split
		}
	}

	if bestIdx < 0 || bestDelta < minDeltaMs {
		return Changepoint{}, false
	}

	before := mean(durations[:bestIdx])
	after := mean(durations[bestIdx:])
	return Changepoint{
		CheckName: name,
		At:        es[bestIdx].Timestamp,
		Before:    before,
		After:     after,
		Delta:     after - before,
	}, true
}

func mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}
