// Package history provides facilities for persisting, querying, and analysing
// historical check results produced by hivecheck.
//
// Anomaly Detection
//
// The anomaly sub-feature scans a sequence of [Entry] values (oldest-first)
// for a named check and classifies notable patterns:
//
//   - Spike     – a single non-OK result surrounded on both sides by OK results,
//                 indicating a transient blip rather than a real outage.
//
//   - Degraded  – three or more consecutive non-OK results, signalling a
//                 sustained service degradation that warrants attention.
//
//   - Recovery  – a transition back to OK following a degraded run, useful for
//                 generating "all-clear" notifications via the alert dispatcher.
//
// Usage:
//
//	entries, _ := store.Load("my-service", 50)
//	anomalies := history.DetectAnomalies("my-service", entries)
//	for _, a := range anomalies {
//		fmt.Println(a)
//	}
package history
