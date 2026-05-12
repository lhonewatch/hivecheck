// Package history provides utilities for storing, analysing, and reporting
// on historical health-check results produced by hivecheck.
//
// # Scorecard
//
// The scorecard module computes a composite health score (0–100) for each
// monitored check using three weighted components:
//
//   - Availability  – percentage of results with StatusOK (default weight 50%)
//   - Reliability   – inverse of the critical-result rate (default weight 30%)
//   - Performance   – normalised inverse of average check duration relative to
//     a configurable worst-case ceiling (default weight 20%)
//
// Usage:
//
//	entries, _ := store.Load("mycheck")
//	scores := history.ComputeScorecard(entries, history.ScorecardOptions{
//		Since:       time.Now().Add(-24 * time.Hour),
//		MaxDuration: 10 * time.Second,
//	})
//	for _, s := range scores {
//		fmt.Println(s)
//	}
//
// Results are sorted in descending order of composite score so the
// healthiest checks appear first.
package history
