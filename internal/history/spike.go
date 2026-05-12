package history

import (
	"fmt"
	"time"
)

// SpikeOptions controls spike detection behaviour.
type SpikeOptions struct {
	// MinSamples is the minimum number of entries required before detection runs.
	MinSamples int
	// ZScoreThreshold is the number of standard deviations above the mean that
	// constitutes a spike in duration.
	ZScoreThreshold float64
	// Window restricts analysis to entries within this duration. Zero means all.
	Window time.Duration
}

// DefaultSpikeOptions returns sensible defaults for spike detection.
func DefaultSpikeOptions() SpikeOptions {
	return SpikeOptions{
		MinSamples:      5,
		ZScoreThreshold: 2.5,
		Window:          24 * time.Hour,
	}
}

// SpikeResult holds the outcome of spike detection for a single check.
type SpikeResult struct {
	CheckName string
	IsSpike   bool
	ZScore    float64
	MeanMs    float64
	StdDevMs  float64
	LatestMs  float64
}

// String returns a human-readable summary of the spike result.
func (s SpikeResult) String() string {
	if !s.IsSpike {
		return fmt.Sprintf("%s: no spike (z=%.2f, mean=%.1fms)", s.CheckName, s.ZScore, s.MeanMs)
	}
	return fmt.Sprintf("%s: SPIKE detected (z=%.2f, latest=%.1fms, mean=%.1fms ±%.1fms)",
		s.CheckName, s.ZScore, s.LatestMs, s.MeanMs, s.StdDevMs)
}

// DetectSpikes analyses historical entries for latency spikes using a
// z-score approach. It returns one SpikeResult per check name found in
// the provided entries.
func DetectSpikes(entries []Entry, opts SpikeOptions) []SpikeResult {
	if len(entries) == 0 {
		return nil
	}

	cutoff := time.Now().Add(-opts.Window)
	byCheck := make(map[string][]float64)

	for _, e := range entries {
		if opts.Window > 0 && e.Timestamp.Before(cutoff) {
			continue
		}
		if e.DurationMs <= 0 {
			continue
		}
		byCheck[e.CheckName] = append(byCheck[e.CheckName], float64(e.DurationMs))
	}

	results := make([]SpikeResult, 0, len(byCheck))
	for name, vals := range byCheck {
		if len(vals) < opts.MinSamples {
			continue
		}
		m := meanFloat(vals)
		sd := stddev(vals, m)
		latest := vals[len(vals)-1]
		z := 0.0
		if sd > 0 {
			z = (latest - m) / sd
		}
		results = append(results, SpikeResult{
			CheckName: name,
			IsSpike:   z >= opts.ZScoreThreshold,
			ZScore:    z,
			MeanMs:    m,
			StdDevMs:  sd,
			LatestMs:  latest,
		})
	}
	return results
}

func meanFloat(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func stddev(vals []float64, m float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	var variance float64
	for _, v := range vals {
		d := v - m
		variance += d * d
	}
	variance /= float64(len(vals))
	return mathSqrt(variance)
}
