package history

import (
	"fmt"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// SLAReport summarises availability and compliance for a single check
// over a given window.
type SLAReport struct {
	CheckName    string
	Window       time.Duration
	Total        int
	OKCount      int
	WarnCount    int
	CritCount    int
	ErrorCount   int
	Availability float64 // percentage of OK results
	Compliant    bool    // true when availability >= target
	Target       float64 // e.g. 99.9
}

// String returns a human-readable one-line summary.
func (r SLAReport) String() string {
	status := "NON-COMPLIANT"
	if r.Compliant {
		status = "COMPLIANT"
	}
	return fmt.Sprintf("%s: %.2f%% availability over %s [target %.2f%%] — %s",
		r.CheckName, r.Availability, r.Window.String(), r.Target, status)
}

// ComputeSLA calculates the SLA report for each check found in entries.
// Only entries within the given window (relative to now) are considered.
// target is the minimum required availability percentage (0–100).
func ComputeSLA(entries []Entry, window time.Duration, target float64) []SLAReport {
	if len(entries) == 0 {
		return nil
	}

	cutoff := time.Now().UTC().Add(-window)

	type counts struct {
		ok, warn, crit, errn int
	}
	byCheck := make(map[string]*counts)

	for _, e := range entries {
		if e.Timestamp.Before(cutoff) {
			continue
		}
		c, ok := byCheck[e.CheckName]
		if !ok {
			c = &counts{}
			byCheck[e.CheckName] = c
		}
		switch e.Status {
		case check.StatusOK:
			c.ok++
		case check.StatusWarn:
			c.warn++
		case check.StatusCritical:
			c.crit++
		default:
			c.errn++
		}
	}

	reports := make([]SLAReport, 0, len(byCheck))
	for name, c := range byCheck {
		total := c.ok + c.warn + c.crit + c.errn
		var avail float64
		if total > 0 {
			avail = float64(c.ok) / float64(total) * 100.0
		}
		reports = append(reports, SLAReport{
			CheckName:    name,
			Window:       window,
			Total:        total,
			OKCount:      c.ok,
			WarnCount:    c.warn,
			CritCount:    c.crit,
			ErrorCount:   c.errn,
			Availability: avail,
			Compliant:    avail >= target,
			Target:       target,
		})
	}
	return reports
}
