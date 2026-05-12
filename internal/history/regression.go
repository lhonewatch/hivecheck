package history

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// RegressionResult holds the outcome of a regression detection analysis for a
// single check. A regression is defined as a statistically significant increase
// in failure rate or average duration compared to a reference window.
type RegressionResult struct {
	CheckName      string
	BaselineRate   float64 // failure rate in the reference window [0,1]
	CurrentRate    float64 // failure rate in the recent window [0,1]
	BaselineAvgMs  float64 // average duration (ms) in the reference window
	CurrentAvgMs   float64 // average duration (ms) in the recent window
	RateRegressed  bool
	LatencyRegressed bool
	Severity       string // "none", "minor", "major"
}

// String returns a human-readable summary of the regression result.
func (r RegressionResult) String() string {
	return fmt.Sprintf(
		"check=%s severity=%s rate=%.2f%%→%.2f%% latency=%.1fms→%.1fms",
		r.CheckName, r.Severity,
		r.BaselineRate*100, r.CurrentRate*100,
		r.BaselineAvgMs, r.CurrentAvgMs,
	)
}

// RegressionOptions controls the detection thresholds.
type RegressionOptions struct {
	// ReferenceWindow is how far back the baseline period extends from the
	// boundary between reference and recent windows.
	ReferenceWindow time.Duration
	// RecentWindow is the size of the window being evaluated.
	RecentWindow time.Duration
	// RateThreshold is the minimum absolute increase in failure rate to flag.
	RateThreshold float64
	// LatencyThresholdPct is the minimum % increase in avg latency to flag.
	LatencyThresholdPct float64
}

// DefaultRegressionOptions returns sensible defaults.
func DefaultRegressionOptions() RegressionOptions {
	return RegressionOptions{
		ReferenceWindow:     24 * time.Hour,
		RecentWindow:        1 * time.Hour,
		RateThreshold:       0.10,
		LatencyThresholdPct: 25.0,
	}
}

// DetectRegressions compares a recent window against a reference baseline for
// each check found in entries and returns one RegressionResult per check.
func DetectRegressions(entries []Entry, opts RegressionOptions) []RegressionResult {
	if len(entries) == 0 {
		return nil
	}

	now := entries[0].Timestamp
	for _, e := range entries {
		if e.Timestamp.After(now) {
			now = e.Timestamp
		}
	}

	recentStart := now.Add(-opts.RecentWindow)
	refStart := recentStart.Add(-opts.ReferenceWindow)

	type bucket struct {
		fails, total int
		durSum       float64
	}

	ref := map[string]*bucket{}
	recent := map[string]*bucket{}

	for _, e := range entries {
		if e.Timestamp.Before(refStart) || !e.Timestamp.Before(recentStart) && e.Timestamp.Before(recentStart) {
			continue
		}
		var b map[string]*bucket
		if !e.Timestamp.Before(recentStart) {
			b = recent
		} else if !e.Timestamp.Before(refStart) {
			b = ref
		} else {
			continue
		}
		if b[e.CheckName] == nil {
			b[e.CheckName] = &bucket{}
		}
		b[e.CheckName].total++
		if e.Status > StatusOK {
			b[e.CheckName].fails++
		}
		b[e.CheckName].durSum += float64(e.DurationMs)
	}

	names := make([]string, 0)
	seen := map[string]bool{}
	for n := range ref {
		if !seen[n] {
			names = append(names, n)
			seen[n] = true
		}
	}
	for n := range recent {
		if !seen[n] {
			names = append(names, n)
			seen[n] = true
		}
	}
	sort.Strings(names)

	results := make([]RegressionResult, 0, len(names))
	for _, name := range names {
		rb := ref[name]
		cb := recent[name]
		if rb == nil || rb.total == 0 || cb == nil || cb.total == 0 {
			continue
		}
		baseRate := float64(rb.fails) / float64(rb.total)
		curRate := float64(cb.fails) / float64(cb.total)
		baseAvg := rb.durSum / float64(rb.total)
		curAvg := cb.durSum / float64(cb.total)

		rateDelta := curRate - baseRate
		latPct := 0.0
		if baseAvg > 0 {
			latPct = (curAvg - baseAvg) / baseAvg * 100
		}

		rateReg := rateDelta >= opts.RateThreshold
		latReg := latPct >= opts.LatencyThresholdPct

		severity := "none"
		switch {
		case rateReg && rateDelta >= 0.30 || latReg && latPct >= 75:
			severity = "major"
		case rateReg || latReg:
			severity = "minor"
		}

		results = append(results, RegressionResult{
			CheckName:        name,
			BaselineRate:     math.Round(baseRate*1000) / 1000,
			CurrentRate:      math.Round(curRate*1000) / 1000,
			BaselineAvgMs:    math.Round(baseAvg*10) / 10,
			CurrentAvgMs:     math.Round(curAvg*10) / 10,
			RateRegressed:    rateReg,
			LatencyRegressed: latReg,
			Severity:         severity,
		})
	}
	return results
}
