// Package history provides utilities for persisting and analysing historical
// check results produced by hivecheck.
//
// # Spike Detection
//
// DetectSpikes identifies latency spikes in a series of historical check
// entries using a z-score approach. For each check name present in the
// supplied entries the function computes the mean and standard deviation of
// all recorded durations within the configured time window, then compares
// the most-recent observation against that baseline.
//
// A result is classified as a spike when its z-score meets or exceeds the
// configured ZScoreThreshold (default 2.5 standard deviations).
//
// Usage:
//
//	opts := history.DefaultSpikeOptions()
//	opts.ZScoreThreshold = 3.0
//	results := history.DetectSpikes(entries, opts)
//	for _, r := range results {
//		fmt.Println(r)
//	}
package history
