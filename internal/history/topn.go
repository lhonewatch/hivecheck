package history

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// TopNEntry holds aggregated statistics for a single check used in top-N ranking.
type TopNEntry struct {
	CheckName   string
	TotalChecks int
	Failures    int
	AvgDuration time.Duration
	FailureRate float64 // 0.0–1.0
}

// TopNResult is the ranked output of ComputeTopN.
type TopNResult struct {
	Ranked []TopNEntry
	Metric string
}

// TopNMetric selects the ranking criterion.
type TopNMetric string

const (
	TopNByFailureRate TopNMetric = "failure_rate"
	TopNByAvgDuration TopNMetric = "avg_duration"
)

// ComputeTopN returns the top n checks ranked by the given metric.
// Entries with zero total checks are excluded.
func ComputeTopN(entries []StoreEntry, n int, metric TopNMetric) TopNResult {
	if n <= 0 {
		n = 5
	}

	agg := map[string]*TopNEntry{}
	for _, e := range entries {
		a, ok := agg[e.CheckName]
		if !ok {
			a = &TopNEntry{CheckName: e.CheckName}
			agg[e.CheckName] = a
		}
		a.TotalChecks++
		if e.Status != StatusOK {
			a.Failures++
		}
		a.AvgDuration += e.Duration
	}

	result := make([]TopNEntry, 0, len(agg))
	for _, a := range agg {
		if a.TotalChecks == 0 {
			continue
		}
		a.AvgDuration /= time.Duration(a.TotalChecks)
		a.FailureRate = float64(a.Failures) / float64(a.TotalChecks)
		result = append(result, *a)
	}

	sort.Slice(result, func(i, j int) bool {
		switch metric {
		case TopNByAvgDuration:
			return result[i].AvgDuration > result[j].AvgDuration
		default: // TopNByFailureRate
			if result[i].FailureRate != result[j].FailureRate {
				return result[i].FailureRate > result[j].FailureRate
			}
			return result[i].CheckName < result[j].CheckName
		}
	})

	if n > len(result) {
		n = len(result)
	}
	return TopNResult{Ranked: result[:n], Metric: string(metric)}
}

// Format returns a human-readable table of the top-N result.
func (r TopNResult) Format() string {
	if len(r.Ranked) == 0 {
		return "top-n: no data"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Top %d checks by %s:\n", len(r.Ranked), r.Metric))
	sb.WriteString(fmt.Sprintf("  %-30s %8s %8s %12s\n", "check", "total", "fail%", "avg_dur"))
	for _, e := range r.Ranked {
		sb.WriteString(fmt.Sprintf("  %-30s %8d %7.1f%% %12s\n",
			e.CheckName, e.TotalChecks, e.FailureRate*100, e.AvgDuration.Round(time.Millisecond)))
	}
	return sb.String()
}
