// Package history — forecast.go
//
// ForecastChecks performs a lightweight linear-regression forecast over
// historical check entries to predict future latency for each named check.
//
// Usage:
//
//	entries, _ := store.Load("my-check")
//	results := history.ForecastChecks(entries, 1*time.Hour)
//	for _, r := range results {
//		fmt.Println(r)
//	}
//
// Confidence levels:
//
//	"low"    — fewer than 7 samples
//	"medium" — 7–19 samples
//	"high"   — 20 or more samples
//
// A ForecastResult is marked Degrading when the regression slope is
// positive, indicating that average duration is increasing over time.
package history
