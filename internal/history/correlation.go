package history

import (
	"fmt"
	"math"
	"sort"
)

// CorrelationResult holds the Pearson correlation coefficient between two checks.
type CorrelationResult struct {
	CheckA      string
	CheckB      string
	Coefficient float64
	SampleSize  int
}

// String returns a human-readable summary of the correlation result.
func (c CorrelationResult) String() string {
	strength := correlationLabel(c.Coefficient)
	return fmt.Sprintf("%s <-> %s: r=%.3f (%s, n=%d)",
		c.CheckA, c.CheckB, c.Coefficient, strength, c.SampleSize)
}

// CorrelateChecks computes Pearson correlation coefficients for all pairs of
// checks found in entries, using the numeric status value as the variable.
// Only pairs with at least minSamples overlapping timestamps are included.
func CorrelateChecks(entries []Entry, minSamples int) []CorrelationResult {
	if len(entries) == 0 || minSamples < 2 {
		return nil
	}

	// Group status values by check name, keyed by truncated-minute timestamp.
	type key struct {
		name string
		t    int64
	}
	seriesMap := make(map[string]map[int64]float64)
	for _, e := range entries {
		ts := e.Timestamp.Unix() / 60 // bucket to minute
		if seriesMap[e.CheckName] == nil {
			seriesMap[e.CheckName] = make(map[int64]float64)
		}
		seriesMap[e.CheckName][ts] = float64(e.Status)
	}

	names := make([]string, 0, len(seriesMap))
	for n := range seriesMap {
		names = append(names, n)
	}
	sort.Strings(names)

	var results []CorrelationResult
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			a, b := names[i], names[j]
			r, n := pearson(seriesMap[a], seriesMap[b])
			if n >= minSamples {
				results = append(results, CorrelationResult{
					CheckA:      a,
					CheckB:      b,
					Coefficient: r,
					SampleSize:  n,
				})
			}
		}
	}
	return results
}

func pearson(a, b map[int64]float64) (float64, int) {
	var xs, ys []float64
	for ts, va := range a {
		if vb, ok := b[ts]; ok {
			xs = append(xs, va)
			ys = append(ys, vb)
		}
	}
	n := len(xs)
	if n < 2 {
		return 0, n
	}
	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := 0; i < n; i++ {
		sumX += xs[i]
		sumY += ys[i]
		sumXY += xs[i] * ys[i]
		sumX2 += xs[i] * xs[i]
		sumY2 += ys[i] * ys[i]
	}
	fn := float64(n)
	num := sumXY - (sumX*sumY)/fn
	den := math.Sqrt((sumX2 - sumX*sumX/fn) * (sumY2 - sumY*sumY/fn))
	if den == 0 {
		return 0, n
	}
	return num / den, n
}

func correlationLabel(r float64) string {
	abs := math.Abs(r)
	switch {
	case abs >= 0.9:
		return "very strong"
	case abs >= 0.7:
		return "strong"
	case abs >= 0.5:
		return "moderate"
	case abs >= 0.3:
		return "weak"
	default:
		return "negligible"
	}
}
