// Package history — incident detection.
//
// DetectIncidents scans a time-ordered slice of history entries and
// identifies contiguous windows where a named check was in a non-OK
// state.  Each window is returned as an Incident value that records:
//
//   - CheckName   – the name of the affected check
//   - StartedAt   – timestamp of the first non-OK entry
//   - ResolvedAt  – timestamp of the first subsequent OK entry
//                   (zero if the check has not yet recovered)
//   - PeakStatus  – the worst status observed during the window
//   - Duration    – elapsed time from start to resolution (or last
//                   observed non-OK entry for open incidents)
//
// Usage:
//
//	entries, _ := store.Load("api-health")
//	incidents := history.DetectIncidents(entries)
//	for _, inc := range incidents {
//		fmt.Println(inc)
//	}
package history
