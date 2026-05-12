package history

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// CapacityReport holds a per-check capacity projection derived from recent
// duration trends. It estimates how many additional load units the check can
// absorb before breaching the critical threshold.
type CapacityReport struct {
	CheckName       string
	AvgDuration     time.Duration
	TrendSlope      float64 // ms per sample
	Projected30     time.Duration // projected avg duration in 30 samples
	HeadroomPct     float64       // % headroom before CriticalCeiling
	CriticalCeiling time.Duration
}

// String returns a human-readable summary of the capacity report.
func (c CapacityReport) String() string {
	return fmt.Sprintf(
		"[%s] avg=%v slope=%.2fms/sample proj30=%v headroom=%.1f%%",
		c.CheckName, c.AvgDuration, c.TrendSlope, c.Projected30, c.HeadroomPct,
	)
}

// CapacityOptions configures ComputeCapacity.
type CapacityOptions struct {
	// CriticalCeiling is the duration beyond which a check is considered
	// critically degraded. Defaults to 5 seconds.
	CriticalCeiling time.Duration
	// MinSamples is the minimum number of entries required. Defaults to 5.
	MinSamples int
}

func defaultCapacityOptions() CapacityOptions {
	return CapacityOptions{
		CriticalCeiling: 5 * time.Second,
		MinSamples:      5,
	}
}

// ComputeCapacity analyses historical entries and returns a CapacityReport for
// each check that has sufficient samples. Entries are grouped by check name and
// a linear regression over duration is used to project future load headroom.
func ComputeCapacity(entries []Entry, opts *CapacityOptions) []CapacityReport {
	if len(entries) == 0 {
		return nil
	}

	o := defaultCapacityOptions()
	if opts != nil {
		if opts.CriticalCeiling > 0 {
			o.CriticalCeiling = opts.CriticalCeiling
		}
		if opts.MinSamples > 0 {
			o.MinSamples = opts.MinSamples
		}
	}

	byCheck := make(map[string][]Entry)
	for _, e := range entries {
		byCheck[e.CheckName] = append(byCheck[e.CheckName], e)
	}

	var reports []CapacityReport
	for name, es := range byCheck {
		if len(es) < o.MinSamples {
			continue
		}
		sort.Slice(es, func(i, j int) bool { return es[i].Timestamp.Before(es[j].Timestamp) })

		durations := make([]float64, len(es))
		var sum float64
		for i, e := range es {
			ms := float64(e.Duration.Milliseconds())
			durations[i] = ms
			sum += ms
		}
		avg := sum / float64(len(durations))

		slope, _ := linearRegression(durations)
		projected := avg + slope*30
		if projected < 0 {
			projected = 0
		}

		ceiling := float64(o.CriticalCeiling.Milliseconds())
		headroom := 0.0
		if ceiling > 0 {
			headroom = math.Max(0, (ceiling-projected)/ceiling*100)
		}

		reports = append(reports, CapacityReport{
			CheckName:       name,
			AvgDuration:     time.Duration(avg) * time.Millisecond,
			TrendSlope:      slope,
			Projected30:     time.Duration(projected) * time.Millisecond,
			HeadroomPct:     headroom,
			CriticalCeiling: o.CriticalCeiling,
		})
	}

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].HeadroomPct < reports[j].HeadroomPct
	})
	return reports
}
