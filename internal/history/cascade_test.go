package history

import (
	"testing"
	"time"
)

func buildCascadeEntries(base time.Time, specs []struct {
	offset time.Duration
	name   string
	status Status
}) []Entry {
	out := make([]Entry, len(specs))
	for i, s := range specs {
		out[i] = Entry{
			Timestamp: base.Add(s.offset),
			CheckName: s.name,
			Status:    s.status,
		}
	}
	return out
}

func TestDetectCascades_Empty(t *testing.T) {
	events, err := DetectCascades(nil, DefaultCascadeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no events, got %d", len(events))
	}
}

func TestDetectCascades_InvalidOptions(t *testing.T) {
	opts := DefaultCascadeOptions()
	opts.WindowDuration = 0
	_, err := DetectCascades(nil, opts)
	if err == nil {
		t.Fatal("expected error for zero WindowDuration")
	}

	opts2 := DefaultCascadeOptions()
	opts2.MinChecks = 1
	_, err = DetectCascades(nil, opts2)
	if err == nil {
		t.Fatal("expected error for MinChecks < 2")
	}
}

func TestDetectCascades_AllOK(t *testing.T) {
	base := time.Now()
	entries := buildCascadeEntries(base, []struct {
		offset time.Duration
		name   string
		status Status
	}{
		{0, "svc-a", StatusOK},
		{10 * time.Second, "svc-b", StatusOK},
		{20 * time.Second, "svc-c", StatusOK},
	})
	events, err := DetectCascades(entries, DefaultCascadeOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no cascade events, got %d", len(events))
	}
}

func TestDetectCascades_Detected(t *testing.T) {
	base := time.Now()
	specs := []struct {
		offset time.Duration
		name   string
		status Status
	}{
		{0, "svc-a", StatusCritical},
		{15 * time.Second, "svc-b", StatusWarning},
		{30 * time.Second, "svc-c", StatusCritical},
		{10 * time.Minute, "svc-d", StatusCritical}, // outside window
	}
	entries := buildCascadeEntries(base, specs)
	opts := DefaultCascadeOptions()
	opts.MinChecks = 3
	events, err := DetectCascades(entries, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 cascade event, got %d", len(events))
	}
	if len(events[0].Checks) != 3 {
		t.Errorf("expected 3 checks in cascade, got %d", len(events[0].Checks))
	}
	if events[0].Checks[0] != "svc-a" {
		t.Errorf("expected root cause svc-a, got %s", events[0].Checks[0])
	}
}

func TestCascadeEvent_String(t *testing.T) {
	base := time.Now()
	e := CascadeEvent{
		Start:  base,
		End:    base.Add(45 * time.Second),
		Checks: []string{"svc-a", "svc-b", "svc-c"},
	}
	s := e.String()
	for _, want := range []string{"cascade", "svc-a", "svc-b", "svc-c", "→"} {
		if !containsStr(s, want) {
			t.Errorf("String() missing %q: %s", want, s)
		}
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && (
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}()))
}
