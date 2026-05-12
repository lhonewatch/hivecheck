// Package history provides utilities for persisting, analysing, and
// summarising historical check results produced by hivecheck.
//
// # Streak Analysis
//
// ComputeStreaks inspects a slice of [Entry] values (ordered oldest-first) and
// returns the current trailing streak for every check present in the data.
//
// A streak is a consecutive sequence of entries that share the same [check.Status].
// When the status changes the streak counter resets, so only the most-recent
// run is reported — useful for alerting on prolonged degradation or for
// celebrating extended healthy periods.
//
// Example:
//
//	entries, _ := store.Load("my-service")
//	streaks := history.ComputeStreaks(entries)
//	for _, s := range streaks {
//		fmt.Println(s)
//	}
package history
