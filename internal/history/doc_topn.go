// Package history provides historical storage, analysis, and reporting
// utilities for hivecheck health-check results.
//
// # Top-N Analysis
//
// ComputeTopN ranks checks by a chosen metric over a slice of StoreEntry
// values. Two metrics are supported:
//
//   - TopNByFailureRate – ranks checks with the highest ratio of non-OK
//     results first. Useful for identifying the most unreliable services.
//
//   - TopNByAvgDuration – ranks checks with the longest average execution
//     time first. Useful for spotting performance regressions.
//
// Example:
//
//	entries, _ := store.Load("my-check")
//	result := history.ComputeTopN(entries, 5, history.TopNByFailureRate)
//	fmt.Print(result.Format())
//
// The Format method returns a plain-text table suitable for CLI output.
package history
