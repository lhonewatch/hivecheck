// Package history provides utilities for storing, analysing, and
// reporting on health-check history.
//
// # Dependency Graph
//
// The dependency graph feature builds a directed graph of check
// relationships inferred from historical correlation data. Checks
// whose duration time-series are strongly correlated are linked as
// potential dependencies, giving operators a visual overview of which
// services tend to fail together.
//
// Usage:
//
//	graph := history.BuildDependencyGraph(entries, history.DependencyOptions{
//		MinCorrelation: 0.75,
//		MinSamples:     10,
//	})
//	fmt.Println(history.FormatDependencyGraph(graph))
package history
