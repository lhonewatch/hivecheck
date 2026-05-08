// Package history provides utilities for persisting, analysing, and
// summarising health-check results over time.
//
// # Sliding Window Analysis
//
// The window sub-feature aggregates raw history entries within a rolling
// time window (1 hour, 24 hours, or 7 days) relative to a reference
// timestamp, typically time.Now().
//
// Usage:
//
//	entries, _ := store.Load("api")
//	stats := history.ComputeWindow(entries, history.WindowDay, time.Now())
//	for _, s := range stats {
//		fmt.Println(s)
//	}
//
// WindowStats exposes per-check counts (OK / Warn / Critical / Error),
// average check duration, and an uptime percentage calculated as the
// proportion of OK results within the window.
package history
