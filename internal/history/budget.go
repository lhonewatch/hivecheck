package history

import (
	"fmt"
	"time"
)

// ErrorBudget represents the remaining error budget for a check over a window.
type ErrorBudget struct {
	CheckName      string
	WindowStart    time.Time
	WindowEnd      time.Time
	TargetPct      float64 // e.g. 99.9
	ActualPct      float64
	BudgetSeconds  float64 // total allowed downtime seconds
	BurnedSeconds  float64 // downtime actually consumed
	RemainingPct   float64 // fraction of budget remaining (0–1)
	Exhausted      bool
}

// String returns a human-readable summary of the error budget.
func (b ErrorBudget) String() string {
	status := "OK"
	if b.Exhausted {
		status = "EXHAUSTED"
	}
	return fmt.Sprintf(
		"[%s] budget=%s actual=%.4f%% target=%.4f%% burned=%.1fs remaining=%.1fs status=%s",
		b.CheckName,
		b.WindowEnd.Sub(b.WindowStart).Round(time.Second),
		b.ActualPct,
		b.TargetPct,
		b.BurnedSeconds,
		b.BudgetSeconds-b.BurnedSeconds,
		status,
	)
}

// ComputeErrorBudget calculates the error budget consumption for each check
// found in entries over the given window, relative to targetPct (e.g. 99.9).
// Only entries within [start, end] are considered.
func ComputeErrorBudget(entries []Entry, start, end time.Time, targetPct float64) []ErrorBudget {
	if len(entries) == 0 || targetPct <= 0 || targetPct >= 100 {
		return nil
	}

	windowSecs := end.Sub(start).Seconds()
	if windowSecs <= 0 {
		return nil
	}

	allowedDownPct := 1.0 - targetPct/100.0
	budgetSecs := windowSecs * allowedDownPct

	type acc struct {
		total  int
		nonOK  int
		burned float64
	}
	byCheck := map[string]*acc{}

	for _, e := range entries {
		if e.Timestamp.Before(start) || e.Timestamp.After(end) {
			continue
		}
		a, ok := byCheck[e.CheckName]
		if !ok {
			a = &acc{}
			byCheck[e.CheckName] = a
		}
		a.total++
		if e.Status != 0 { // non-OK
			a.nonOK++
			a.burned += e.Duration.Seconds()
		}
	}

	results := make([]ErrorBudget, 0, len(byCheck))
	for name, a := range byCheck {
		var actualPct float64
		if a.total > 0 {
			actualPct = float64(a.total-a.nonOK) / float64(a.total) * 100.0
		}
		remaining := 1.0
		if budgetSecs > 0 {
			remaining = 1.0 - a.burned/budgetSecs
		}
		if remaining < 0 {
			remaining = 0
		}
		results = append(results, ErrorBudget{
			CheckName:     name,
			WindowStart:   start,
			WindowEnd:     end,
			TargetPct:     targetPct,
			ActualPct:     actualPct,
			BudgetSeconds: budgetSecs,
			BurnedSeconds: a.burned,
			RemainingPct:  remaining,
			Exhausted:     a.burned >= budgetSecs,
		})
	}
	return results
}
