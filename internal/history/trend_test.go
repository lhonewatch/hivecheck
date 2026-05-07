package history_test

import (
	"strings"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
	"github.com/example/hivecheck/internal/history"
)

func buildEntries(statuses []check.Status) []history.Entry {
	entries := make([]history.Entry, len(statuses))
	for i, s := range statuses {
		entries[i] = history.Entry{
			Timestamp: time.Now(),
			Overall:   s,
			Results:   []check.Result{{Name: "svc", Status: s}},
		}
	}
	return entries
}

func TestAnalyse_Empty(t *testing.T) {
	if trends := history.Analyse(nil); trends != nil {
		t.Fatalf("expected nil, got %v", trends)
	}
}

func TestAnalyse_StableOK(t *testing.T) {
	entries := buildEntries([]check.Status{
		check.StatusOK, check.StatusOK, check.StatusOK,
	})
	trends := history.Analyse(entries)
	if len(trends) != 1 {
		t.Fatalf("expected 1 trend, got %d", len(trends))
	}
	if trends[0].Flapping {
		t.Error("stable check should not be flapping")
	}
	if trends[0].LastStatus != check.StatusOK {
		t.Errorf("expected OK, got %s", trends[0].LastStatus)
	}
}

func TestAnalyse_Flapping(t *testing.T) {
	entries := buildEntries([]check.Status{
		check.StatusOK, check.StatusCritical, check.StatusOK,
		check.StatusCritical, check.StatusOK,
	})
	trends := history.Analyse(entries)
	if !trends[0].Flapping {
		t.Error("expected flapping=true")
	}
}

func TestFormatTrend_ContainsName(t *testing.T) {
	t0 := history.Trend{
		CheckName:    "my-service",
		StatusCounts: map[string]int{"ok": 3},
		LastStatus:   check.StatusOK,
		Flapping:     false,
	}
	line := history.FormatTrend(t0)
	if !strings.Contains(line, "my-service") {
		t.Errorf("FormatTrend missing name: %q", line)
	}
}

func TestFormatTrend_FlappingLabel(t *testing.T) {
	t0 := history.Trend{
		CheckName:  "unstable",
		LastStatus: check.StatusWarn,
		Flapping:   true,
	}
	if !strings.Contains(history.FormatTrend(t0), "FLAPPING") {
		t.Error("expected FLAPPING label in output")
	}
}
