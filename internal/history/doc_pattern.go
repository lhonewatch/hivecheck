// Package history provides historical check result storage,
// analysis, and reporting for hivecheck.
//
// # Pattern Detection
//
// DetectPatterns analyses a slice of history entries and identifies
// recurring temporal patterns in check behaviour — for example, a
// service that consistently degrades at the same hour of the day.
//
// Usage:
//
//	patterns := history.DetectPatterns(entries, history.DefaultPatternOptions())
//	for _, p := range patterns {
//		fmt.Println(p.String())
//	}
//
// Each PatternResult describes the check name, the detected pattern
// kind (HourOfDay, DayOfWeek), the bucket that fires most often, and
// a confidence score in [0, 1].
package history
