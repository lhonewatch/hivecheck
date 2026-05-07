// Package history provides persistent storage and trend analysis for
// hivecheck run results.
//
// # Overview
//
// Each time hivecheck runs it can persist a history.Entry to a local
// directory via Store.Save. Previously saved entries can be loaded with
// Store.Load and passed to Analyse to compute per-check Trend values that
// indicate whether a service is stable, degraded, or flapping.
//
// # Usage
//
//	store, err := history.NewStore(".hivecheck/history")
//	if err != nil { ... }
//
//	entry := history.Entry{
//		Timestamp: time.Now(),
//		Results:   results,
//		Overall:   overall,
//	}
//	if err := store.Save(entry); err != nil { ... }
//
//	entries, err := store.Load()
//	trends := history.Analyse(entries)
//	for _, t := range trends {
//		fmt.Println(history.FormatTrend(t))
//	}
package history
