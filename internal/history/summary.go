package history

import (
	"fmt"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// Entry represents a persisted snapshot of a single check result.
type Entry struct {
	CheckName string        `json:"check_name"`
	Status    check.Status  `json:"status"`
	Message   string        `json:"message"`
	Value     float64       `json:"value"`
	Timestamp time.Time     `json:"timestamp"`
}

// Summary aggregates historical entries for a named check.
type Summary struct {
	CheckName   string
	Total       int
	OKCount     int
	WarnCount   int
	CritCount   int
	ErrorCount  int
	LastStatus  check.Status
	LastChecked time.Time
}

// Summarise builds a Summary from a slice of historical entries for one check.
// Entries must all share the same CheckName; only the latest timestamp is used
// for LastChecked.
func Summarise(entries []Entry) (Summary, error) {
	if len(entries) == 0 {
		return Summary{}, fmt.Errorf("history: no entries provided")
	}

	s := Summary{
		CheckName: entries[0].CheckName,
		Total:     len(entries),
	}

	var latest time.Time

	for _, e := range entries {
		switch e.Status {
		case check.StatusOK:
			s.OKCount++
		case check.StatusWarn:
			s.WarnCount++
		case check.StatusCritical:
			s.CritCount++
		default:
			s.ErrorCount++
		}
		if e.Timestamp.After(latest) {
			latest = e.Timestamp
			s.LastStatus = e.Status
			s.LastChecked = e.Timestamp
		}
	}

	return s, nil
}

// Format returns a human-readable single-line representation of the Summary.
func (s Summary) Format() string {
	return fmt.Sprintf(
		"[%s] last=%s ok=%d warn=%d crit=%d err=%d (total=%d, checked=%s)",
		s.CheckName,
		s.LastStatus,
		s.OKCount,
		s.WarnCount,
		s.CritCount,
		s.ErrorCount,
		s.Total,
		s.LastChecked.Format(time.RFC3339),
	)
}
