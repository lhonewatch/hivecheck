package history

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// DigestPeriod defines the time window for a digest report.
type DigestPeriod int

const (
	DigestDaily  DigestPeriod = iota // last 24 hours
	DigestWeekly                     // last 7 days
)

// DigestEntry summarises a single check within a digest.
type DigestEntry struct {
	CheckName    string
	TotalRuns    int
	OKCount      int
	WarnCount    int
	CritCount    int
	ErrorCount   int
	AvgDurationMs float64
	WorstStatus  string
}

// Digest holds the full digest report for a period.
type Digest struct {
	Period    DigestPeriod
	GeneratedAt time.Time
	Entries   []DigestEntry
}

// BuildDigest computes a digest over stored entries within the given period.
// entries should be pre-loaded from the Store.
func BuildDigest(entries []Entry, period DigestPeriod, now time.Time) Digest {
	var cutoff time.Time
	switch period {
	case DigestWeekly:
		cutoff = now.Add(-7 * 24 * time.Hour)
	default:
		cutoff = now.Add(-24 * time.Hour)
	}

	type agg struct {
		ok, warn, crit, errCount int
		totalDuration            float64
		worst                    int // numeric severity
		worstStatus              string
	}

	severity := map[string]int{"ok": 0, "warn": 1, "critical": 2, "error": 3}

	buckets := make(map[string]*agg)
	for _, e := range entries {
		if e.Timestamp.Before(cutoff) {
			continue
		}
		a, ok := buckets[e.CheckName]
		if !ok {
			a = &agg{worstStatus: "ok"}
			buckets[e.CheckName] = a
		}
		switch strings.ToLower(e.Status) {
		case "ok":
			a.ok++
		case "warn":
			a.warn++
		case "critical":
			a.crit++
		default:
			a.errCount++
		}
		a.totalDuration += float64(e.DurationMs)
		if sev, found := severity[strings.ToLower(e.Status)]; found && sev > a.worst {
			a.worst = sev
			a.worstStatus = e.Status
		}
	}

	names := make([]string, 0, len(buckets))
	for n := range buckets {
		names = append(names, n)
	}
	sort.Strings(names)

	result := Digest{Period: period, GeneratedAt: now}
	for _, name := range names {
		a := buckets[name]
		total := a.ok + a.warn + a.crit + a.errCount
		avg := 0.0
		if total > 0 {
			avg = a.totalDuration / float64(total)
		}
		result.Entries = append(result.Entries, DigestEntry{
			CheckName:     name,
			TotalRuns:     total,
			OKCount:       a.ok,
			WarnCount:     a.warn,
			CritCount:     a.crit,
			ErrorCount:    a.errCount,
			AvgDurationMs: avg,
			WorstStatus:   a.worstStatus,
		})
	}
	return result
}

// Format returns a human-readable summary of the digest.
func (d Digest) Format() string {
	var sb strings.Builder
	label := "Daily"
	if d.Period == DigestWeekly {
		label = "Weekly"
	}
	fmt.Fprintf(&sb, "%s Digest (%s)\n", label, d.GeneratedAt.UTC().Format("2006-01-02 15:04 UTC"))
	fmt.Fprintf(&sb, "%-24s %6s %4s %4s %4s %4s %10s %s\n",
		"Check", "Runs", "OK", "Warn", "Crit", "Err", "AvgMs", "Worst")
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 72))
	for _, e := range d.Entries {
		fmt.Fprintf(&sb, "%-24s %6d %4d %4d %4d %4d %10.1f %s\n",
			e.CheckName, e.TotalRuns, e.OKCount, e.WarnCount,
			e.CritCount, e.ErrorCount, e.AvgDurationMs, e.WorstStatus)
	}
	return sb.String()
}
