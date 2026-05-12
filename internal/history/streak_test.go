package history

import (
	"strings"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
)

func buildStreakEntries(name string, statuses []check.Status, base time.Time) []Entry {
	ent := make([]Entry, len(statuses))
	for i, s := range statuses {
		ent[i] = Entry{
			CheckName: name,
			Status:    s,
			Timestamp: base.Add(time.Duration(i) * time.Minute),
		}
	}
	return ent
}

func TestComputeStreaks_Empty(t *testing.T) {
	result := ComputeStreaks(nil)
	if len(result) != 0 {
		t.Fatalf("expected empty, got %d", len(result))
	}
}

func TestComputeStreaks_AllSameStatus(t *testing.T) {
	base := time.Now()
	entries := buildStreakEntries("svc-a", []check.Status{
		check.StatusOK, check.StatusOK, check.StatusOK,
	}, base)

	streaks := ComputeStreaks(entries)
	if len(streaks) != 1 {
		t.Fatalf("expected 1 streak, got %d", len(streaks))
	}
	s := streaks[0]
	if s.Length != 3 {
		t.Errorf("expected length 3, got %d", s.Length)
	}
	if s.Status != check.StatusOK {
		t.Errorf("expected OK, got %s", s.Status)
	}
}

func TestComputeStreaks_BreakResetsStreak(t *testing.T) {
	base := time.Now()
	entries := buildStreakEntries("svc-b", []check.Status{
		check.StatusOK, check.StatusOK,
		check.StatusCritical, check.StatusCritical, check.StatusCritical,
	}, base)

	streaks := ComputeStreaks(entries)
	if len(streaks) != 1 {
		t.Fatalf("expected 1 streak per check, got %d", len(streaks))
	}
	s := streaks[0]
	// trailing streak is Critical with length 3
	if s.Status != check.StatusCritical {
		t.Errorf("expected Critical, got %s", s.Status)
	}
	if s.Length != 3 {
		t.Errorf("expected length 3, got %d", s.Length)
	}
}

func TestComputeStreaks_MultipleChecks(t *testing.T) {
	base := time.Now()
	a := buildStreakEntries("alpha", []check.Status{check.StatusOK, check.StatusOK}, base)
	b := buildStreakEntries("beta", []check.Status{check.StatusWarn, check.StatusWarn, check.StatusWarn}, base)

	streaks := ComputeStreaks(append(a, b...))
	if len(streaks) != 2 {
		t.Fatalf("expected 2 streaks, got %d", len(streaks))
	}
	byName := make(map[string]Streak)
	for _, s := range streaks {
		byName[s.CheckName] = s
	}
	if byName["alpha"].Length != 2 {
		t.Errorf("alpha length: want 2, got %d", byName["alpha"].Length)
	}
	if byName["beta"].Length != 3 {
		t.Errorf("beta length: want 3, got %d", byName["beta"].Length)
	}
}

func TestStreak_String_ContainsFields(t *testing.T) {
	s := Streak{
		CheckName: "my-service",
		Status:    check.StatusWarn,
		Length:    5,
		Since:     time.Now().Add(-10 * time.Minute),
		Until:     time.Now(),
		Duration:  10 * time.Minute,
	}
	out := s.String()
	for _, want := range []string{"my-service", "warn", "5"} {
		if !strings.Contains(out, want) {
			t.Errorf("String() missing %q in %q", want, out)
		}
	}
}
