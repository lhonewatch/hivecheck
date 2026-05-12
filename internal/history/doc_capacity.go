// Package history provides utilities for storing, analysing, and reporting
// on historical health-check results collected by hivecheck.
//
// # Capacity Planning
//
// The capacity sub-system estimates how much headroom a service has before
// its response-time or failure-rate metrics breach configured thresholds.
// It does so by fitting a linear trend to the most-recent window of history
// entries and extrapolating forward.
//
// # Usage
//
//	opts := history.defaultCapacityOptions()
//	opts.Horizon = 48 * time.Hour
//	reports := history.ComputeCapacity(entries, opts)
//	for _, r := range reports {
//		fmt.Println(r.String())
//	}
//
// A CapacityReport is produced for every check that has sufficient samples.
// Checks with fewer samples than opts.MinSamples are skipped.
package history
