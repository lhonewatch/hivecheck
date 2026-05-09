// Package history provides utilities for storing, analysing, and reporting
// on historical health-check results.
//
// # Error Budget
//
// ComputeErrorBudget calculates SLO error-budget consumption for each check
// over a caller-supplied time window and availability target.
//
// Given a target availability percentage (e.g. 99.9 %), it derives the total
// allowed downtime (budget) for the window and compares it against the
// cumulative duration of non-OK results to determine how much budget has been
// burned and whether the budget is exhausted.
//
// Usage:
//
//	budgets := history.ComputeErrorBudget(entries, start, end, 99.9)
//	for _, b := range budgets {
//		fmt.Println(b)
//	}
package history
