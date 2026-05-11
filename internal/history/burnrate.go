package history

import (
	"fmt"
	"math"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// BurnRateOptions configures the burn-rate computation.
type BurnRateOptions struct {
	// SLOPercent is the target availability percentage, e.g. 99.9.
	SLOPercent float64
	// WindowHours is the look-back window in hours (e.g. 1 for a 1-hour window).
	WindowHours int
}

// BurnRateReport holds burn-rate results for a single check.
type BurnRateReport struct {
	CheckName   string
	BurnRate    float64 // relative to SLO window
	ErrorCount  int
	TotalCount  int
	WindowStart time.Time
	WindowEnd   time.Time
}

// String returns a human-readable summary of the burn-rate report.
func (r BurnRateReport) String() string {
	return fmt.Sprintf(
		"[%s] burn_rate=%.2f errors=%d/%d window=%s–%s",
		r.CheckName,
		r.BurnRate,
		r.ErrorCount,
		r.TotalCount,
		r.WindowStart.Format(time.RFC3339),
		r.WindowEnd.Format(time.RFC3339),
	)
}

// ComputeBurnRate calculates the per-check burn rate for the given entries
// within the look-back window defined by opts.
// A burn rate of 1.0 equals the steady-state error rate allowed by the SLO.
func ComputeBurnRate(entries []StoreEntry, opts BurnRateOptions) []BurnRateReport {
	if len(entries) == 0 || opts.WindowHours <= 0 || opts.SLOPercent <= 0 {
		return nil
	}

	sloErrorRate := 1.0 - (opts.SLOPercent / 100.0)
	if sloErrorRate <= 0 {
		sloErrorRate = math.SmallestNonzeroFloat64
	}

	now := time.Now().UTC()
	cutoff := now.Add(-time.Duration(opts.WindowHours) * time.Hour)

	type counts struct {
		total  int
		errors int
		start  time.Time
	}
	agg := make(map[string]*counts)

	for _, e := range entries {
		if e.Timestamp.Before(cutoff) {
			continue
		}
		c, ok := agg[e.CheckName]
		if !ok {
			c = &counts{start: e.Timestamp}
			agg[e.CheckName] = c
		}
		c.total++
		if e.Status == check.StatusCritical || e.Status == check.StatusError {
			c.errors++
		}
		if e.Timestamp.Before(c.start) {
			c.start = e.Timestamp
		}
	}

	var reports []BurnRateReport
	for name, c := range agg {
		if c.total == 0 {
			continue
		}
		observedRate := float64(c.errors) / float64(c.total)
		burnRate := observedRate / sloErrorRate
		reports = append(reports, BurnRateReport{
			CheckName:   name,
			BurnRate:    burnRate,
			ErrorCount:  c.errors,
			TotalCount:  c.total,
			WindowStart: c.start,
			WindowEnd:   now,
		})
	}
	return reports
}
