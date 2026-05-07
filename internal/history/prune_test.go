package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeEntry(t *testing.T, dir, filename string, modTime time.Time) {
	t.Helper()
	path := filepath.Join(dir, filename)
	data, _ := json.Marshal(map[string]string{"check": "test"})
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("writeEntry: %v", err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
}

func TestPrune_MaxEntries(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()

	for i := 0; i < 5; i++ {
		writeEntry(t, dir, filepath.Base(tempEntryName("svc", now.Add(time.Duration(i)*time.Minute))), now)
	}

	deleted, err := Prune(dir, PruneOptions{MaxEntries: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 2 {
		t.Errorf("expected 2 deleted, got %d", deleted)
	}

	remaining, _ := os.ReadDir(dir)
	if len(remaining) != 3 {
		t.Errorf("expected 3 remaining files, got %d", len(remaining))
	}
}

func TestPrune_MaxAge(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	old := now.Add(-48 * time.Hour)

	writeEntry(t, dir, "mysvc-old.json", old)
	writeEntry(t, dir, "mysvc-new.json", now)

	deleted, err := Prune(dir, PruneOptions{MaxAge: 24 * time.Hour})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}
}

func TestPrune_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	deleted, err := Prune(dir, PruneOptions{MaxEntries: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
}

func TestPrune_NonExistentDir(t *testing.T) {
	deleted, err := Prune("/tmp/hivecheck-nonexistent-xyz", PruneOptions{MaxEntries: 2})
	if err != nil {
		t.Fatalf("non-existent dir should not error: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
}

func TestCheckNameFromFile(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"my-service-20240101T120000.json", "my-service"},
		{"svc-20240101.json", "svc"},
		{"notseparated.json", "notseparated"},
	}
	for _, tc := range cases {
		got := checkNameFromFile(tc.input)
		if got != tc.want {
			t.Errorf("checkNameFromFile(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// tempEntryName mirrors the naming used by Store so prune can group correctly.
func tempEntryName(checkName string, ts time.Time) string {
	return checkName + "-" + ts.UTC().Format("20060102T150405") + ".json"
}
