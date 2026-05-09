package history

import (
	"fmt"
	"time"
)

// WindowSize defines the granularity of a sliding window.
type WindowSize int

const (
	WindowHour  WindowSize = iota // last 60 minutes
	WindowDay                      // last 24 hours
	WindowWeek                     // last 7 days
)

// WindowStats holds aggregated statistics for a single check over a window.
type WindowStats struct {
	CheckName   string
	Window      WindowSize
	Total       int
	OKCount     int
	WarnCount   int
	CritCount   int
	ErrorCount  int
	AvgDuration time.Duration
	UptimePct   float64
}

// String returns a human-readable summary of the window stats.
func (w WindowStats) String() string {
	windowLabel := map[WindowSize]string{
		WindowHour: "1h",
		WindowDay:  "24h",
		WindowWeek: "7d",
	}[w.Window]
	return fmt.Sprintf(
		"[%s] %s — total=%d ok=%d warn=%d crit=%d err=%d uptime=%.1f%% avg=%s",
		windowLabel, w.CheckName,
		w.Total, w.OKCount, w.WarnCount, w.CritCount, w.ErrorCount,
		w.UptimePct, w.AvgDuration.Round(time.Millisecond),
	)
}

// IsHealthy reports whether the window stats are considered healthy, defined
// as having no critical or error results and an uptime of at least minUptime
// percent (e.g. pass 99.0 for 99%).
func (w WindowStats) IsHealthy(minUptime float64) bool {
	if w.CritCount > 0 || w.ErrorCount > 0 {
		return false
	}
	return w.Total == 0 || w.UptimePct >= minUptime
}

// ComputeWindow aggregates entries for each check within the given window
// relative to now. Only entries whose Timestamp falls within [now-window, now]
// are included.
func ComputeWindow(entries []Entry, size WindowSize, now time.Time) []WindowStats {
	duration := windowDuration(size)
	cutoff := now.Add(-duration)

	type acc struct {
		total, ok, warn, crit, errCount int
		totalDur                        time.Duration
	}
	byCheck := make(map[string]*acc)

	for _, e := range entries {
		if e.Timestamp.Before(cutoff) || e.Timestamp.After(now) {
			continue
		}
		a, exists := byCheck[e.CheckName]
		if !exists {
			a = &acc{}
			byCheck[e.CheckName] = a
		}
		a.total++
		a.totalDur += e.Duration
		switch e.Status {
		case StatusOK:
			a.ok++
		case StatusWarn:
			a.warn++
		case StatusCritical:
			a.crit++
		default:
			a.errCount++
		}
	}

	results := make([]WindowStats, 0, len(byCheck))
	for name, a := range byCheck {
		var avg time.Duration
		if a.total > 0 {
			avg = a.totalDur / time.Duration(a.total)
		}
		uptime := 0.0
		if a.total > 0 {
			uptime = float64(a.ok) / float64(a.total) * 100.0
		}
		results = append(results, WindowStats{
			CheckName:   name,
			Window:      size,
			Total:       a.total,
			OKCount:     a.ok,
			WarnCount:   a.warn,
			CritCount:   a.crit,
			ErrorCount:  a.errCount,
			AvgDuration: avg,
			UptimePct:   uptime,
		})
	}
	return results
}

func windowDuration(size WindowSize) time.Duration {
	switch size {
	case WindowHour:
		return time.Hour
	case WindowDay:
		return 24 * time.Hour
	case WindowWeek:
		return 7 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}
