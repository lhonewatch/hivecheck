// Package report provides formatting and summarization utilities for
// HiveCheck run results.
//
// # Summary
//
// A [Summary] aggregates all [check.Result] values produced by a single
// runner invocation together with timing metadata. It exposes helpers such
// as [Summary.Overall] (worst observed status) and [Summary.StatusCounts]
// (per-status histogram).
//
// # Formatters
//
// Two built-in [Formatter] implementations are provided:
//
//   - [TextFormatter] – human-readable plain-text output suitable for
//     terminal display.
//   - [JSONFormatter] – machine-readable JSON, optionally indented, suitable
//     for piping to other tools or storing as artefacts.
//
// Custom formatters can be added by implementing the [Formatter] interface.
package report
