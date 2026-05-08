package history

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// HeatmapCell represents the aggregated status for a single check in a given time bucket.
type HeatmapCell struct {
	CheckName  string
	Bucket     time.Time
	OKCount    int
	WarnCount  int
	CritCount  int
	ErrorCount int
	Total      int
}

// DominantStatus returns the most frequent non-OK status, or "ok" if all checks passed.
func (c HeatmapCell) DominantStatus() string {
	switch {
	case c.ErrorCount > 0 && c.ErrorCount >= c.CritCount && c.ErrorCount >= c.WarnCount:
		return "error"
	case c.CritCount > 0 && c.CritCount >= c.WarnCount:
		return "critical"
	case c.WarnCount > 0:
		return "warn"
	default:
		return "ok"
	}
}

// BuildHeatmap aggregates history entries into a grid of cells bucketed by the
// given duration (e.g. time.Hour for hourly, 24*time.Hour for daily).
// Only entries within [from, to) are included.
func BuildHeatmap(entries []Entry, bucketSize time.Duration, from, to time.Time) []HeatmapCell {
	type key struct {
		name   string
		bucket time.Time
	}
	cells := make(map[key]*HeatmapCell)

	for _, e := range entries {
		if e.Timestamp.Before(from) || !e.Timestamp.Before(to) {
			continue
		}
		bucket := e.Timestamp.Truncate(bucketSize)
		k := key{name: e.CheckName, bucket: bucket}
		if _, ok := cells[k]; !ok {
			cells[k] = &HeatmapCell{CheckName: e.CheckName, Bucket: bucket}
		}
		c := cells[k]
		c.Total++
		switch strings.ToLower(e.Status) {
		case "ok":
			c.OKCount++
		case "warn":
			c.WarnCount++
		case "critical":
			c.CritCount++
		default:
			c.ErrorCount++
		}
	}

	result := make([]HeatmapCell, 0, len(cells))
	for _, c := range cells {
		result = append(result, *c)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].CheckName != result[j].CheckName {
			return result[i].CheckName < result[j].CheckName
		}
		return result[i].Bucket.Before(result[j].Bucket)
	})
	return result
}

// FormatHeatmap renders a compact text table of the heatmap cells.
func FormatHeatmap(cells []HeatmapCell, bucketFmt string) string {
	if len(cells) == 0 {
		return "no heatmap data"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-24s %-18s %s\n", "check", "bucket", "dominant"))
	for _, c := range cells {
		sb.WriteString(fmt.Sprintf("%-24s %-18s %s (%d/%d ok)\n",
			c.CheckName,
			c.Bucket.Format(bucketFmt),
			c.DominantStatus(),
			c.OKCount,
			c.Total,
		))
	}
	return sb.String()
}
