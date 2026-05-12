// Package history provides facilities for storing, querying, and analysing
// historical health-check results.
//
// # Velocity
//
// ComputeVelocity measures the rate of change of failure rate and average
// check duration across a sliding time window. It is useful for detecting
// services that are degrading quickly rather than just those that are already
// in a bad state.
//
// Each VelocityResult reports:
//   - FailureRatePerHour – how fast the fraction of failing checks is rising
//     (negative values indicate recovery).
//   - DurationDeltaMs    – how fast average latency is changing per hour.
//   - Accelerating       – true when the second half of the window shows a
//     higher failure velocity than the first half.
//
// A minimum of 4 samples per check is required; checks with fewer samples are
// silently skipped.
package history
