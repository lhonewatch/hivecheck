package check

import (
	"fmt"
	"time"
)

// Status represents the health check result level.
type Status int

const (
	StatusOK Status = iota
	StatusWarn
	StatusCritical
	StatusUnknown
)

func (s Status) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusWarn:
		return "WARN"
	case StatusCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Threshold defines warn/critical boundaries for a numeric metric.
type Threshold struct {
	Warn     float64 `yaml:"warn"`
	Critical float64 `yaml:"critical"`
}

// Evaluate returns the Status for a given value against the threshold.
// Values at or above Critical are CRITICAL, at or above Warn are WARN.
func (t Threshold) Evaluate(value float64) Status {
	if value >= t.Critical {
		return StatusCritical
	}
	if value >= t.Warn {
		return StatusWarn
	}
	return StatusOK
}

// Result holds the outcome of a single health check execution.
type Result struct {
	Name      string
	Status    Status
	Message   string
	Value     float64
	Threshold Threshold
	Duration  time.Duration
	Err       error
}

func (r Result) String() string {
	if r.Err != nil {
		return fmt.Sprintf("[%s] %s — error: %v", r.Status, r.Name, r.Err)
	}
	return fmt.Sprintf("[%s] %s — %s (value=%.2f, warn=%.2f, critical=%.2f, duration=%s)",
		r.Status, r.Name, r.Message, r.Value, r.Threshold.Warn, r.Threshold.Critical, r.Duration)
}
