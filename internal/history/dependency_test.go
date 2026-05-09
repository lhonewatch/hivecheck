package history

import (
	"strings"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
)

func buildDepEntries(t *testing.T) []HistoryEntry {
	t.Helper()
	base := time.Now().Add(-30 * time.Minute)
	var entries []HistoryEntry
	for i := 0; i < 15; i++ {
		d := time.Duration(i+1) * 10 * time.Millisecond
		entries = append(entries,
			HistoryEntry{
				Timestamp: base.Add(time.Duration(i) * time.Minute),
				Result: check.Result{
					Name:     "alpha",
					Status:   check.StatusOK,
					Duration: d,
				},
			},
			HistoryEntry{
				Timestamp: base.Add(time.Duration(i) * time.Minute),
				Result: check.Result{
					Name:     "beta",
					Status:   check.StatusOK,
					Duration: d + 5*time.Millisecond, // near-perfect correlation
				},
			},
		)
	}
	return entries
}

func TestBuildDependencyGraph_Empty(t *testing.T) {
	g := BuildDependencyGraph(nil, DefaultDependencyOptions())
	if len(g.Edges) != 0 {
		t.Fatalf("expected 0 edges, got %d", len(g.Edges))
	}
}

func TestBuildDependencyGraph_StrongCorrelation(t *testing.T) {
	entries := buildDepEntries(t)
	opts := DependencyOptions{MinCorrelation: 0.90, MinSamples: 5}
	g := BuildDependencyGraph(entries, opts)
	if len(g.Edges) == 0 {
		t.Fatal("expected at least one edge for strongly correlated checks")
	}
	e := g.Edges[0]
	if e.From != "alpha" || e.To != "beta" {
		t.Errorf("unexpected edge %s->%s", e.From, e.To)
	}
	if e.Correlation < 0.90 {
		t.Errorf("correlation %.2f below threshold", e.Correlation)
	}
}

func TestBuildDependencyGraph_BelowThreshold(t *testing.T) {
	entries := buildDepEntries(t)
	// Require perfect correlation — should produce no edges.
	opts := DependencyOptions{MinCorrelation: 1.0, MinSamples: 5}
	g := BuildDependencyGraph(entries, opts)
	if len(g.Edges) != 0 {
		t.Fatalf("expected 0 edges at threshold 1.0, got %d", len(g.Edges))
	}
}

func TestBuildDependencyGraph_NoDuplicateEdges(t *testing.T) {
	entries := buildDepEntries(t)
	opts := DependencyOptions{MinCorrelation: 0.5, MinSamples: 3}
	g := BuildDependencyGraph(entries, opts)
	seen := make(map[string]int)
	for _, e := range g.Edges {
		key := e.From + "->" + e.To
		seen[key]++
		if seen[key] > 1 {
			t.Errorf("duplicate edge: %s", key)
		}
	}
}

func TestFormatDependencyGraph_Empty(t *testing.T) {
	out := FormatDependencyGraph(DependencyGraph{})
	if !strings.Contains(out, "no significant") {
		t.Errorf("expected no-correlations message, got: %s", out)
	}
}

func TestFormatDependencyGraph_ContainsEdge(t *testing.T) {
	g := DependencyGraph{
		Edges: []DependencyEdge{
			{From: "svc-a", To: "svc-b", Correlation: 0.95, Label: "strong positive"},
		},
	}
	out := FormatDependencyGraph(g)
	for _, want := range []string{"svc-a", "svc-b", "0.95", "strong positive"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q: %s", want, out)
		}
	}
}
