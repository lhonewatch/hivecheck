package history_test

import (
	"os"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
	"github.com/example/hivecheck/internal/history"
)

func makeEntry(name string, status check.Status) history.Entry {
	return history.Entry{
		Timestamp: time.Now().UTC(),
		Overall:   status,
		Results: []check.Result{
			{Name: name, Status: status, Message: "ok"},
		},
	}
}

func TestStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store, err := history.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	e1 := makeEntry("svc-a", check.StatusOK)
	e2 := makeEntry("svc-b", check.StatusWarn)

	for _, e := range []history.Entry{e1, e2} {
		if err := store.Save(e); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	entries, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestStore_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	store, _ := history.NewStore(dir)
	entries, err := store.Load()
	if err != nil {
		t.Fatalf("Load empty: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestNewStore_CreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := base + "/nested/history"
	_, err := history.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore nested: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatal("directory was not created")
	}
}
