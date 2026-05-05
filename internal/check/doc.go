// Package check provides the core health check primitives for hivecheck.
//
// It defines:
//
//   - Threshold: warn/critical numeric boundaries for a metric.
//   - Status: the result level (OK, WARN, CRITICAL, UNKNOWN).
//   - Result: the outcome of a single check execution, including value,
//     status, message, duration, and any error.
//   - Definition: pairs a named CheckFunc with its Threshold configuration.
//   - Runner: executes multiple Definitions concurrently within a timeout
//     and returns a slice of Results.
//
// Typical usage:
//
//	defs := []check.Definition{
//		{
//			Name:      "response_time_ms",
//			Threshold: check.Threshold{Warn: 200, Critical: 500},
//			Fn: myLatencyCheck,
//		},
//	}
//	runner := check.NewRunner(10*time.Second, defs)
//	results := runner.Run(ctx)
//	for _, r := range results {
//		fmt.Println(r)
//	}
package check
