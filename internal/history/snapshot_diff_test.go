package history_test

import (
	"strings"
	"testing"

	"github.com/example/hivecheck/internal/check"
	"github.com/example/hivecheck/internal/history"
)

func snapWith(results []check.Result) history.Snapshot {
	return history.NewSnapshot(results)
}

func TestDiffSnapshots_NoChange(t *testing.T) {
	results := []check.Result{{Name: "api", Status: check.StatusOK}}
	diffs := history.DiffSnapshots(snapWith(results), snapWith(results))
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d", len(diffs))
	}
}

func TestDiffSnapshots_StatusChange(t *testing.T) {
	prev := []check.Result{{Name: "api", Status: check.StatusOK}}
	curr := []check.Result{{Name: "api", Status: check.StatusCritical}}
	diffs := history.DiffSnapshots(snapWith(prev), snapWith(curr))
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if !diffs[0].Changed() {
		t.Error("expected Changed() == true")
	}
	if diffs[0].CheckName != "api" {
		t.Errorf("unexpected check name: %s", diffs[0].CheckName)
	}
}

func TestDiffSnapshots_NewCheck(t *testing.T) {
	prev := []check.Result{}
	curr := []check.Result{{Name: "db", Status: check.StatusWarn}}
	diffs := history.DiffSnapshots(snapWith(prev), snapWith(curr))
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff for new check, got %d", len(diffs))
	}
}

func TestDiffSnapshots_RemovedCheck(t *testing.T) {
	prev := []check.Result{{Name: "cache", Status: check.StatusOK}}
	curr := []check.Result{}
	diffs := history.DiffSnapshots(snapWith(prev), snapWith(curr))
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff for removed check, got %d", len(diffs))
	}
}

func TestFormatDiff_Empty(t *testing.T) {
	out := history.FormatDiff(nil)
	if out != "no status changes detected" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestFormatDiff_ContainsName(t *testing.T) {
	diffs := []history.DiffEntry{{CheckName: "payments", Prev: check.StatusOK, Curr: check.StatusCritical}}
	out := history.FormatDiff(diffs)
	if !strings.Contains(out, "payments") {
		t.Errorf("expected output to contain check name, got: %q", out)
	}
}
