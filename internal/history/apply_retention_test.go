package history

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestApplyRetention_InvalidPolicy(t *testing.T) {
	s, _ := NewStore(t.TempDir())
	err := ApplyRetention(s, RetentionPolicy{MaxEntries: -1}, io.Discard)
	if err == nil {
		t.Fatal("expected error for invalid policy")
	}
}

func TestApplyRetention_ZeroPolicy(t *testing.T) {
	s, _ := NewStore(t.TempDir())
	var buf strings.Builder
	err := ApplyRetention(s, RetentionPolicy{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "skipping") {
		t.Errorf("expected skip message, got: %s", buf.String())
	}
}

func TestApplyRetention_NothingToPrune(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	var buf strings.Builder
	err := ApplyRetention(s, RetentionPolicy{MaxEntries: 10}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "nothing") {
		t.Errorf("expected nothing-to-prune message, got: %s", buf.String())
	}
}

func TestApplyRetention_PrunesOldEntries(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)

	// Write 3 entries older than 1 hour.
	old := time.Now().Add(-2 * time.Hour)
	for i := 0; i < 3; i++ {
		name := filepath.Join(dir, "svc-check_"+old.Add(time.Duration(i)*time.Second).Format("20060102T150405")+".json")
		data := `{"check":"svc-check","status":"ok","value":1.0,"ts":"` + old.Format(time.RFC3339) + `"}`
		_ = os.WriteFile(name, []byte(data), 0644)
	}

	var buf strings.Builder
	err := ApplyRetention(s, RetentionPolicy{MaxAge: time.Hour}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "pruned") {
		t.Errorf("expected pruned message, got: %s", buf.String())
	}
}

func TestPluralSuffix(t *testing.T) {
	if pluralSuffix(1) != "y" {
		t.Error("expected 'y' for 1")
	}
	if pluralSuffix(2) != "ies" {
		t.Error("expected 'ies' for 2")
	}
}
