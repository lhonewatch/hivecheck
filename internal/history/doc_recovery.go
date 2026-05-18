// Package history provides utilities for persisting, analysing, and
// reporting on historical health-check results.
//
// # Recovery Detection
//
// DetectRecoveries scans a slice of history entries and identifies
// transitions where a check moves from a non-OK status (Warn or Critical)
// back to OK. For each such transition it records:
//
//   - The check name
//   - The timestamp at which the failure began
//   - The timestamp at which recovery was observed
//   - The total downtime duration
//   - The status the check held while failing
//
// The returned RecoveryReport aggregates all events and computes the
// average downtime across all recovery events, making it easy to
// understand typical recovery times for each service.
package history
