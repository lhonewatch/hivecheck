// Package history provides utilities for recording, analysing, and
// reporting on historical health-check results.
//
// # Correlation
//
// The correlation sub-feature computes Pearson correlation coefficients
// between pairs of checks over a shared set of time-bucketed observations.
// This helps identify checks whose health tends to rise and fall together,
// which can indicate shared dependencies or cascading failure paths.
//
// Usage:
//
//	results := history.CorrelateChecks(entries, 5)
//	for _, r := range results {
//		fmt.Println(r)
//	}
//
// Timestamps are bucketed to the nearest minute so that checks with
// slightly different execution times are still considered concurrent.
// Only pairs with at least minSamples overlapping buckets are returned.
package history
