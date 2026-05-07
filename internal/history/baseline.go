package history

import (
	"fmt"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// Baseline holds the computed baseline statistics for a single check derived
// from its historical entries.
type Baseline struct {
	CheckName      string
	SampleCount    int
	OKRate         float64 // fraction of samples with status OK
	WarnRate       float64 // fraction of samples with status Warn
	CriticalRate   float64 // fraction of samples with status Critical
	AvgDuration    time.Duration
	ComputedAt     time.Time
}

// String returns a human-readable summary of the baseline.
func (b Baseline) String() string {
	return fmt.Sprintf(
		"Baseline[%s] samples=%d ok=%.1f%% warn=%.1f%% critical=%.1f%% avg_duration=%s",
		b.CheckName,
		b.SampleCount,
		b.OKRate*100,
		b.WarnRate*100,
		b.CriticalRate*100,
		b.AvgDuration.Round(time.Millisecond),
	)
}

// ComputeBaseline calculates a Baseline from the provided history entries for
// the named check. Returns an empty Baseline (SampleCount == 0) when entries
// is nil or empty.
func ComputeBaseline(checkName string, entries []Entry) Baseline {
	b := Baseline{
		CheckName:  checkName,
		ComputedAt: time.Now(),
	}

	if len(entries) == 0 {
		return b
	}

	var totalDur time.Duration
	var okCount, warnCount, critCount int

	for _, e := range entries {
		totalDur += e.Duration
		switch e.Status {
		case check.StatusOK:
			okCount++
		case check.StatusWarn:
			warnCount++
		case check.StatusCritical:
			critCount++
		}
	}

	n := len(entries)
	b.SampleCount = n
	b.OKRate = float64(okCount) / float64(n)
	b.WarnRate = float64(warnCount) / float64(n)
	b.CriticalRate = float64(critCount) / float64(n)
	b.AvgDuration = totalDur / time.Duration(n)

	return b
}
