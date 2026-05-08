// Package history provides facilities for recording, querying, and analysing
// historical check results produced by hivecheck.
//
// # SLA Reporting
//
// ComputeSLA calculates per-check Service Level Agreement (SLA) reports over a
// rolling time window. For each check it counts result statuses, derives an
// availability percentage (ratio of OK results to total results), and compares
// that figure against a caller-supplied target.
//
// Example usage:
//
//	reports := history.ComputeSLA(entries, 7*24*time.Hour, 99.9)
//	for _, r := range reports {
//		fmt.Println(r)
//	}
//
// The SLAReport.String() method returns a concise human-readable summary
// suitable for log output or CLI display.
package history
