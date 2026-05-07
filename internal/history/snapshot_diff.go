package history

import (
	"fmt"
	"strings"

	"github.com/example/hivecheck/internal/check"
)

// DiffEntry describes a status change for a single check between two snapshots.
type DiffEntry struct {
	CheckName string
	Prev      check.Status
	Curr      check.Status
}

// Changed reports whether the status actually changed.
func (d DiffEntry) Changed() bool { return d.Prev != d.Curr }

// DiffSnapshots compares two snapshots and returns entries where the status
// changed. Checks present in only one snapshot are included with StatusUnknown
// (0) as the absent side.
func DiffSnapshots(prev, curr Snapshot) []DiffEntry {
	prevMap := indexByName(prev.Results)
	currMap := indexByName(curr.Results)

	seen := make(map[string]struct{})
	var diffs []DiffEntry

	for name, cr := range currMap {
		seen[name] = struct{}{}
		pr, ok := prevMap[name]
		if !ok {
			diffs = append(diffs, DiffEntry{CheckName: name, Prev: check.StatusOK, Curr: cr.Status})
			continue
		}
		if pr.Status != cr.Status {
			diffs = append(diffs, DiffEntry{CheckName: name, Prev: pr.Status, Curr: cr.Status})
		}
	}
	for name, pr := range prevMap {
		if _, ok := seen[name]; !ok {
			diffs = append(diffs, DiffEntry{CheckName: name, Prev: pr.Status, Curr: check.StatusOK})
		}
	}
	return diffs
}

// FormatDiff returns a human-readable summary of snapshot differences.
func FormatDiff(diffs []DiffEntry) string {
	if len(diffs) == 0 {
		return "no status changes detected"
	}
	var sb strings.Builder
	for _, d := range diffs {
		fmt.Fprintf(&sb, "  %-30s %s -> %s\n", d.CheckName, d.Prev, d.Curr)
	}
	return sb.String()
}

func indexByName(results []check.Result) map[string]check.Result {
	m := make(map[string]check.Result, len(results))
	for _, r := range results {
		m[r.Name] = r
	}
	return m
}
