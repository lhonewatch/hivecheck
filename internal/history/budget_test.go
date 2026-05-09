package history

import (
	"strings"
	"testing"
	"time"
)

func buildBudgetEntries(name string, statuses []int, base time.Time) []Entry {
	entries := make([]Entry, len(statuses))
	for i, s := range statuses {
		entries[i] = Entry{
			CheckName: name,
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Status:    s,
			Duration:  30 * time.Second,
		}
	}
	return entries
}

func TestComputeErrorBudget_Empty(t *testing.T) {
	now := time.Now()
	got := ComputeErrorBudget(nil, now.Add(-time.Hour), now, 99.9)
	if len(got) != 0 {
		t.Fatalf("expected empty, got %d", len(got))
	}
}

func TestComputeErrorBudget_AllOK(t *testing.T) {
	now := time.Now()
	start := now.Add(-time.Hour)
	entries := buildBudgetEntries("api", []int{0, 0, 0, 0, 0}, start)
	got := ComputeErrorBudget(entries, start, now, 99.0)
	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	if got[0].BurnedSeconds != 0 {
		t.Errorf("expected 0 burned, got %.1f", got[0].BurnedSeconds)
	}
	if got[0].Exhausted {
		t.Error("expected budget not exhausted")
	}
	if got[0].ActualPct != 100.0 {
		t.Errorf("expected 100%% actual, got %.2f", got[0].ActualPct)
	}
}

func TestComputeErrorBudget_Exhausted(t *testing.T) {
	now := time.Now()
	start := now.Add(-time.Hour)
	// 4 of 5 checks failing => burned = 4*30s = 120s
	// budget for 99% over 1h = 0.01 * 3600 = 36s — well exceeded
	entries := buildBudgetEntries("db", []int{0, 1, 2, 1, 2}, start)
	got := ComputeErrorBudget(entries, start, now, 99.0)
	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	if !got[0].Exhausted {
		t.Error("expected budget exhausted")
	}
	if got[0].RemainingPct != 0 {
		t.Errorf("expected 0 remaining, got %.4f", got[0].RemainingPct)
	}
}

func TestComputeErrorBudget_MultipleChecks(t *testing.T) {
	now := time.Now()
	start := now.Add(-time.Hour)
	e1 := buildBudgetEntries("svc-a", []int{0, 0, 0}, start)
	e2 := buildBudgetEntries("svc-b", []int{1, 1, 1}, start)
	got := ComputeErrorBudget(append(e1, e2...), start, now, 99.9)
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}
}

func TestComputeErrorBudget_InvalidTarget(t *testing.T) {
	now := time.Now()
	start := now.Add(-time.Hour)
	entries := buildBudgetEntries("x", []int{0}, start)
	if got := ComputeErrorBudget(entries, start, now, 0); len(got) != 0 {
		t.Error("expected nil for targetPct=0")
	}
	if got := ComputeErrorBudget(entries, start, now, 100); len(got) != 0 {
		t.Error("expected nil for targetPct=100")
	}
}

func TestErrorBudget_String_ContainsName(t *testing.T) {
	now := time.Now()
	b := ErrorBudget{
		CheckName:     "my-service",
		WindowStart:   now.Add(-time.Hour),
		WindowEnd:     now,
		TargetPct:     99.9,
		ActualPct:     99.5,
		BudgetSeconds: 3.6,
		BurnedSeconds: 1.8,
		RemainingPct:  0.5,
		Exhausted:     false,
	}
	s := b.String()
	if !strings.Contains(s, "my-service") {
		t.Errorf("String() missing check name: %s", s)
	}
	if !strings.Contains(s, "OK") {
		t.Errorf("String() missing status: %s", s)
	}
}
