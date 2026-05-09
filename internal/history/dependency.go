package history

import (
	"fmt"
	"sort"
	"strings"
)

// DependencyEdge represents a directed relationship between two checks
// inferred from correlated duration metrics.
type DependencyEdge struct {
	From        string
	To          string
	Correlation float64
	Label       string
}

// DependencyGraph holds all inferred edges between checks.
type DependencyGraph struct {
	Edges []DependencyEdge
}

// DependencyOptions controls how the graph is constructed.
type DependencyOptions struct {
	// MinCorrelation is the minimum absolute Pearson correlation required
	// for an edge to be created (0.0–1.0).
	MinCorrelation float64
	// MinSamples is the minimum number of overlapping data points required.
	MinSamples int
}

// DefaultDependencyOptions returns sensible defaults.
func DefaultDependencyOptions() DependencyOptions {
	return DependencyOptions{
		MinCorrelation: 0.70,
		MinSamples:     5,
	}
}

// BuildDependencyGraph analyses entries and returns a DependencyGraph
// whose edges represent strongly correlated check pairs.
func BuildDependencyGraph(entries []HistoryEntry, opts DependencyOptions) DependencyGraph {
	if opts.MinSamples <= 0 {
		opts.MinSamples = 1
	}

	corrs := CorrelateChecks(entries, opts.MinSamples)

	var edges []DependencyEdge
	seen := make(map[string]bool)

	for _, c := range corrs {
		if c.Correlation < opts.MinCorrelation {
			continue
		}
		// Normalise key so A→B and B→A are deduplicated.
		key := c.CheckA + "\x00" + c.CheckB
		if c.CheckA > c.CheckB {
			key = c.CheckB + "\x00" + c.CheckA
		}
		if seen[key] {
			continue
		}
		seen[key] = true

		edges = append(edges, DependencyEdge{
			From:        c.CheckA,
			To:          c.CheckB,
			Correlation: c.Correlation,
			Label:       c.Label,
		})
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})

	return DependencyGraph{Edges: edges}
}

// FormatDependencyGraph returns a human-readable representation of the
// graph suitable for terminal output.
func FormatDependencyGraph(g DependencyGraph) string {
	if len(g.Edges) == 0 {
		return "dependency graph: no significant correlations found\n"
	}
	var sb strings.Builder
	sb.WriteString("dependency graph:\n")
	for _, e := range g.Edges {
		sb.WriteString(fmt.Sprintf("  %s --> %s  (r=%.2f, %s)\n",
			e.From, e.To, e.Correlation, e.Label))
	}
	return sb.String()
}
