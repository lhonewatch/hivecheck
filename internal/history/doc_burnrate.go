// Package history provides facilities for storing, querying, and analysing
// historical health-check results produced by hivecheck.
//
// # Burn-Rate Analysis
//
// The burn-rate sub-module computes how quickly a service is consuming its
// error budget relative to a target SLO window.  A burn rate of 1.0 means
// the service is exhausting its budget at exactly the rate that would deplete
// it by the end of the SLO window; values above 1.0 indicate faster-than-
// expected degradation and should trigger escalating alerts.
//
// Usage:
//
//	report := history.ComputeBurnRate(entries, history.BurnRateOptions{
//		SLOPercent:  99.9,
//		WindowHours: 1,
//	})
//	fmt.Println(report)
package history
