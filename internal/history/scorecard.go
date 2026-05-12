package history

import (
	"fmt"
	"sort"
	"time"

	"github.com/user/hivecheck/internal/check"
)

// ScorecardEntry holds the computed health score for a single check.
type ScorecardEntry struct {
	CheckName    string
	Score        float64 // 0.0–100.0
	Availability float64 // percentage of OK results
	Reliability  float64 // 1 - (critical_rate)
	Perf         float64 // normalised inverse of avg duration
	LastStatus   check.Status
	Samples      int
}

// String returns a one-line summary of the scorecard entry.
func (e ScorecardEntry) String() string {
	return fmt.Sprintf("%s score=%.1f avail=%.1f%% rel=%.1f%% perf=%.1f%% samples=%d last=%s",
		e.CheckName, e.Score,
		e.Availability, e.Reliability, e.Perf,
		e.Samples, e.LastStatus)
}

// ScorecardOptions controls how the scorecard is computed.
type ScorecardOptions struct {
	Since    time.Time
	WeightOK float64 // weight for availability component (default 0.5)
	WeightRel float64 // weight for reliability component (default 0.3)
	WeightPerf float64 // weight for performance component (default 0.2)
	MaxDuration time.Duration // duration treated as worst-case for perf score
}

// ComputeScorecard calculates a composite health score per check from stored
// history entries. Entries older than opts.Since are excluded when non-zero.
func ComputeScorecard(entries []Entry, opts ScorecardOptions) []ScorecardEntry {
	if len(entries) == 0 {
		return nil
	}
	if opts.WeightOK == 0 {
		opts.WeightOK = 0.5
	}
	if opts.WeightRel == 0 {
		opts.WeightRel = 0.3
	}
	if opts.WeightPerf == 0 {
		opts.WeightPerf = 0.2
	}
	if opts.MaxDuration == 0 {
		opts.MaxDuration = 30 * time.Second
	}

	type agg struct {
		ok, warn, crit, total int
		durSum               time.Duration
		last                 check.Status
		lastTime             time.Time
	}
	byName := make(map[string]*agg)

	for _, e := range entries {
		if !opts.Since.IsZero() && e.Timestamp.Before(opts.Since) {
			continue
		}
		a := byName[e.CheckName]
		if a == nil {
			a = &agg{}
			byName[e.CheckName] = a
		}
		a.total++
		a.durSum += e.Duration
		switch e.Status {
		case check.StatusOK:
			a.ok++
		case check.StatusWarn:
			a.warn++
		case check.StatusCritical:
			a.crit++
		}
		if e.Timestamp.After(a.lastTime) {
			a.lastTime = e.Timestamp
			a.last = e.Status
		}
	}

	result := make([]ScorecardEntry, 0, len(byName))
	for name, a := range byName {
		avail := 100.0 * float64(a.ok) / float64(a.total)
		rel := 100.0 * (1.0 - float64(a.crit)/float64(a.total))
		avgDur := a.durSum / time.Duration(a.total)
		perfRatio := 1.0 - float64(avgDur)/float64(opts.MaxDuration)
		if perfRatio < 0 {
			perfRatio = 0
		}
		perf := 100.0 * perfRatio
		score := opts.WeightOK*avail + opts.WeightRel*rel + opts.WeightPerf*perf
		result = append(result, ScorecardEntry{
			CheckName:    name,
			Score:        score,
			Availability: avail,
			Reliability:  rel,
			Perf:         perf,
			LastStatus:   a.last,
			Samples:      a.total,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})
	return result
}
