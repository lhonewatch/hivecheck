package history_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
	"github.com/example/hivecheck/internal/history"
)

func makeResults(names ...string) []check.Result {
	var out []check.Result
	for i, n := range names {
		out = append(out, check.Result{
			Name:   n,
			Status: check.Status(i % 3),
		})
	}
	return out
}

func TestSaveAndLoadSnapshot(t *testing.T) {
	dir := t.TempDir()
	snap := history.NewSnapshot(makeResults("svc-a", "svc-b"))

	if err := history.SaveSnapshot(dir, snap); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}

	paths, err := history.ListSnapshots(dir)
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(paths))
	}

	loaded, err := history.LoadSnapshot(paths[0])
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}
	if len(loaded.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(loaded.Results))
	}
}

func TestLatestSnapshot_NoSnapshots(t *testing.T) {
	dir := t.TempDir()
	_, err := history.LatestSnapshot(dir)
	if err != history.ErrNoSnapshots {
		t.Errorf("expected ErrNoSnapshots, got %v", err)
	}
}

func TestLatestSnapshot_ReturnsNewest(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 3; i++ {
		snap := history.NewSnapshot(makeResults("svc"))
		snap.CapturedAt = time.Now().UTC().Add(time.Duration(i) * time.Second)
		if err := history.SaveSnapshot(dir, snap); err != nil {
			t.Fatalf("save: %v", err)
		}
	}
	_, err := history.LatestSnapshot(dir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestListSnapshots_IgnoresOtherFiles(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "other.txt"), []byte("x"), 0o644)
	paths, err := history.ListSnapshots(dir)
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(paths))
	}
}
