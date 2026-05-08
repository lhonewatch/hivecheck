package history

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// PercentileResult holds latency percentile statistics for a single check.
type PercentileResult struct {
	CheckName string
	P50       time.Duration
	P90       time.Duration
	P95       time.Duration
	P99       time.Duration
	SampleCount int
}

// String returns a human-readable summary of the percentile result.
func (p PercentileResult) String() string {
	return fmt.Sprintf("%s (n=%d): p50=%s p90=%s p95=%s p99=%s",
		p.CheckName, p.SampleCount, p.P50, p.P90, p.P95, p.P99)
}

// ComputePercentiles calculates latency percentiles for each check found in
// the provided entries. Entries with zero duration are excluded from the
// calculation so that missing data does not skew results.
func ComputePercentiles(entries []Entry) []PercentileResult {
	if len(entries) == 0 {
		return nil
	}

	// Group durations by check name.
	groups := make(map[string][]float64)
	for _, e := range entries {
		if e.Duration <= 0 {
			continue
		}
		groups[e.CheckName] = append(groups[e.CheckName], float64(e.Duration))
	}

	results := make([]PercentileResult, 0, len(groups))
	for name, durations := range groups {
		sort.Float64s(durations)
		results = append(results, PercentileResult{
			CheckName:   name,
			P50:         time.Duration(percentile(durations, 50)),
			P90:         time.Duration(percentile(durations, 90)),
			P95:         time.Duration(percentile(durations, 95)),
			P99:         time.Duration(percentile(durations, 99)),
			SampleCount: len(durations),
		})
	}

	// Sort by check name for deterministic output.
	sort.Slice(results, func(i, j int) bool {
		return results[i].CheckName < results[j].CheckName
	})
	return results
}

// percentile returns the p-th percentile value from a pre-sorted slice using
// the nearest-rank method.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p/100.0*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
