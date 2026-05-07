// Package history provides utilities for persisting, analysing, and managing
// the run history of health checks.
//
// # Retention
//
// The RetentionPolicy type describes how many entries to keep per check and
// for how long. Use DefaultRetentionPolicy for a sensible starting point:
//
//	policy := history.DefaultRetentionPolicy()
//	if err := history.ApplyRetention(store, policy, os.Stdout); err != nil {
//	    log.Fatal(err)
//	}
//
// ApplyRetention validates the policy and delegates to Prune. It writes a
// human-readable summary line to the provided io.Writer.
//
// Policies with both MaxEntries and MaxAge set to zero are treated as
// no-ops — no files are removed.
package history
