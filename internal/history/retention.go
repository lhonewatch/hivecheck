package history

import (
	"time"
)

// RetentionPolicy defines rules for how long history entries are kept.
type RetentionPolicy struct {
	// MaxEntries is the maximum number of entries to retain per check.
	// Zero means no limit.
	MaxEntries int

	// MaxAge is the maximum age of an entry before it is pruned.
	// Zero means no age limit.
	MaxAge time.Duration
}

// DefaultRetentionPolicy returns a sensible default policy:
// keep at most 100 entries per check, no older than 30 days.
func DefaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{
		MaxEntries: 100,
		MaxAge:     30 * 24 * time.Hour,
	}
}

// IsZero reports whether the policy has no constraints set.
func (p RetentionPolicy) IsZero() bool {
	return p.MaxEntries == 0 && p.MaxAge == 0
}

// Validate returns an error if the policy contains invalid values.
func (p RetentionPolicy) Validate() error {
	if p.MaxEntries < 0 {
		return &RetentionError{Field: "max_entries", Value: p.MaxEntries}
	}
	if p.MaxAge < 0 {
		return &RetentionError{Field: "max_age", Value: int(p.MaxAge)}
	}
	return nil
}

// RetentionError describes an invalid retention policy field.
type RetentionError struct {
	Field string
	Value int
}

func (e *RetentionError) Error() string {
	return "invalid retention policy: " + e.Field + " must not be negative"
}
