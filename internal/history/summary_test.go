package history

import (
	"strings"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
)

func buildTestEntries(name string, statuses []check.Status) []Entry {
	base := time.Now().Add(-time.Duration(len(statuses)) * time.Minute)
	entries := make([]Entry, len(statuses))
	for i, st := range statuses {
		entries[i] = Entry{
			CheckName: name,
			Status:    st,
			Message:   st.String(),
			Value:     float64(i),
			Timestamp: base.Add(time.Duration(i) * time.Minute),
		}
	}
	return entries
}

func TestSummarise_Empty(t *testing.T) {
	_, err := Summarise(nil)
	if err == nil {
		t.Fatal("expected error for empty entries, got nil")
	}
}

func TestSummarise_Counts(t *testing.T) {
	entries := buildTestEntries("db", []check.Status{
		check.StatusOK,
		check.StatusOK,
		check.StatusWarn,
		check.StatusCritical,
		check.StatusCritical,
	})

	s, err := Summarise(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.CheckName != "db" {
		t.Errorf("CheckName = %q, want %q", s.CheckName, "db")
	}
	if s.Total != 5 {
		t.Errorf("Total = %d, want 5", s.Total)
	}
	if s.OKCount != 2 {
		t.Errorf("OKCount = %d, want 2", s.OKCount)
	}
	if s.WarnCount != 1 {
		t.Errorf("WarnCount = %d, want 1", s.WarnCount)
	}
	if s.CritCount != 2 {
		t.Errorf("CritCount = %d, want 2", s.CritCount)
	}
}

func TestSummarise_LastStatus(t *testing.T) {
	entries := buildTestEntries("api", []check.Status{
		check.StatusCritical,
		check.StatusOK,
		check.StatusWarn,
	})

	s, err := Summarise(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Last entry (highest timestamp) has StatusWarn.
	if s.LastStatus != check.StatusWarn {
		t.Errorf("LastStatus = %v, want %v", s.LastStatus, check.StatusWarn)
	}
}

func TestSummary_Format(t *testing.T) {
	entries := buildTestEntries("cache", []check.Status{check.StatusOK})
	s, _ := Summarise(entries)
	out := s.Format()

	for _, want := range []string{"cache", "ok=", "warn=", "crit=", "total="} {
		if !strings.Contains(out, want) {
			t.Errorf("Format() missing %q in output: %s", want, out)
		}
	}
}
