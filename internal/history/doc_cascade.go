// Package history provides historical check result storage and analysis.
//
// # Cascade Failure Detection
//
// The cascade sub-feature identifies chains of check failures that propagate
// across services within a short time window, suggesting a root-cause event
// that triggered downstream degradation.
//
// Usage:
//
//	opts := history.DefaultCascadeOptions()
//	opts.WindowDuration = 2 * time.Minute
//	results, err := history.DetectCascades(entries, opts)
//	for _, r := range results {
//		fmt.Println(r.String())
//	}
//
// A cascade is reported when three or more distinct checks transition to a
// non-OK status within the configured time window, ordered by first-failure
// timestamp so the likely root cause appears first.
package history
