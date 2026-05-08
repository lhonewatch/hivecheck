package history

import (
	"strings"
	"testing"
	"time"
)

func buildHeatmapEntries(base time.Time) []Entry {
	return []Entry{
		{CheckName: "api", Status: "ok", Timestamp: base},
		{CheckName: "api", Status: "ok", Timestamp: base.Add(10 * time.Minute)},
		{CheckName: "api", Status: "warn", Timestamp: base.Add(20 * time.Minute)},
		{CheckName: "db", Status: "critical", Timestamp: base},
		{CheckName: "db", Status: "critical", Timestamp: base.Add(5 * time.Minute)},
		{CheckName: "db", Status: "ok", Timestamp: base.Add(90 * time.Minute)}, // next bucket
	}
}

func TestBuildHeatmap_Empty(t *testing.T) {
	now := time.Now()
	cells := BuildHeatmap(nil, time.Hour, now.Add(-time.Hour), now)
	if len(cells) != 0 {
		t.Fatalf("expected 0 cells, got %d", len(cells))
	}
}

func TestBuildHeatmap_BucketCount(t *testing.T) {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	entries := buildHeatmapEntries(base)
	from := base
	to := base.Add(3 * time.Hour)

	cells := BuildHeatmap(entries, time.Hour, from, to)
	// api: bucket@10h, db: bucket@10h, db: bucket@11h => 3 cells
	if len(cells) != 3 {
		t.Fatalf("expected 3 cells, got %d", len(cells))
	}
}

func TestBuildHeatmap_DominantStatus(t *testing.T) {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	entries := buildHeatmapEntries(base)
	from := base
	to := base.Add(3 * time.Hour)

	cells := BuildHeatmap(entries, time.Hour, from, to)

	byKey := make(map[string]HeatmapCell)
	for _, c := range cells {
		byKey[c.CheckName+"@"+c.Bucket.Format("15")] = c
	}

	if got := byKey["api@10"].DominantStatus(); got != "warn" {
		t.Errorf("api@10 dominant: want warn, got %s", got)
	}
	if got := byKey["db@10"].DominantStatus(); got != "critical" {
		t.Errorf("db@10 dominant: want critical, got %s", got)
	}
	if got := byKey["db@11"].DominantStatus(); got != "ok" {
		t.Errorf("db@11 dominant: want ok, got %s", got)
	}
}

func TestBuildHeatmap_ExcludesOutOfRange(t *testing.T) {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	entries := buildHeatmapEntries(base)
	// narrow window excludes the 90-minute entry
	from := base
	to := base.Add(time.Hour)

	cells := BuildHeatmap(entries, time.Hour, from, to)
	for _, c := range cells {
		if c.CheckName == "db" && c.Bucket.Hour() == 11 {
			t.Error("should not include db entry in hour 11")
		}
	}
}

func TestFormatHeatmap_Empty(t *testing.T) {
	out := FormatHeatmap(nil, "2006-01-02 15")
	if out != "no heatmap data" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestFormatHeatmap_ContainsFields(t *testing.T) {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	entries := buildHeatmapEntries(base)
	cells := BuildHeatmap(entries, time.Hour, base, base.Add(3*time.Hour))
	out := FormatHeatmap(cells, "2006-01-02 15")

	for _, want := range []string{"api", "db", "ok", "critical", "warn"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
}
