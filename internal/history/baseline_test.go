package history

import (
	"strings"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
)

func buildBaselineEntries(statuses []check.Status, dur time.Duration) []Entry {
	entries := make([]Entry, len(statuses))
	for i, s := range statuses {
		entries[i] = Entry{
			CheckName: "svc",
			Status:    s,
			Duration:  dur,
			Timestamp: time.Now().Add(-time.Duration(i) * time.Minute),
		}
	}
	return entries
}

func TestComputeBaseline_Empty(t *testing.T) {
	b := ComputeBaseline("svc", nil)
	if b.SampleCount != 0 {
		t.Fatalf("expected 0 samples, got %d", b.SampleCount)
	}
	if b.CheckName != "svc" {
		t.Fatalf("expected check name svc, got %s", b.CheckName)
	}
}

func TestComputeBaseline_AllOK(t *testing.T) {
	entries := buildBaselineEntries(
		[]check.Status{check.StatusOK, check.StatusOK, check.StatusOK, check.StatusOK},
		200*time.Millisecond,
	)
	b := ComputeBaseline("svc", entries)

	if b.SampleCount != 4 {
		t.Fatalf("expected 4 samples, got %d", b.SampleCount)
	}
	if b.OKRate != 1.0 {
		t.Errorf("expected OKRate 1.0, got %f", b.OKRate)
	}
	if b.WarnRate != 0 || b.CriticalRate != 0 {
		t.Errorf("unexpected non-zero warn/critical rates")
	}
	if b.AvgDuration != 200*time.Millisecond {
		t.Errorf("expected avg 200ms, got %s", b.AvgDuration)
	}
}

func TestComputeBaseline_Mixed(t *testing.T) {
	entries := buildBaselineEntries(
		[]check.Status{
			check.StatusOK, check.StatusOK,
			check.StatusWarn,
			check.StatusCritical,
		},
		100*time.Millisecond,
	)
	b := ComputeBaseline("svc", entries)

	if b.SampleCount != 4 {
		t.Fatalf("expected 4 samples, got %d", b.SampleCount)
	}
	if got := b.OKRate; got != 0.5 {
		t.Errorf("OKRate: want 0.5, got %f", got)
	}
	if got := b.WarnRate; got != 0.25 {
		t.Errorf("WarnRate: want 0.25, got %f", got)
	}
	if got := b.CriticalRate; got != 0.25 {
		t.Errorf("CriticalRate: want 0.25, got %f", got)
	}
}

func TestBaseline_String_ContainsName(t *testing.T) {
	entries := buildBaselineEntries(
		[]check.Status{check.StatusOK},
		50*time.Millisecond,
	)
	b := ComputeBaseline("my-service", entries)
	s := b.String()
	if !strings.Contains(s, "my-service") {
		t.Errorf("String() missing check name: %s", s)
	}
	if !strings.Contains(s, "samples=1") {
		t.Errorf("String() missing sample count: %s", s)
	}
}
