package history

import (
	"testing"
	"time"
)

func buildClusterEntries() []Entry {
	now := time.Now()
	make := func(name string, dur time.Duration, n int) []Entry {
		out := make([]Entry, n)
		for i := range out {
			out[i] = Entry{CheckName: name, Timestamp: now, Duration: dur}
		}
		return out
	}
	var all []Entry
	all = append(all, make("fast-svc", 10*time.Millisecond, 5)...)
	all = append(all, make("fast-db", 15*time.Millisecond, 5)...)
	all = append(all, make("medium-api", 120*time.Millisecond, 5)...)
	all = append(all, make("slow-worker", 800*time.Millisecond, 5)...)
	return all
}

func TestClusterChecks_Empty(t *testing.T) {
	result := ClusterChecks(nil, 3)
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d clusters", len(result))
	}
}

func TestClusterChecks_ZeroK(t *testing.T) {
	entries := buildClusterEntries()
	result := ClusterChecks(entries, 0)
	if len(result) != 0 {
		t.Fatalf("expected empty result for k=0, got %d", len(result))
	}
}

func TestClusterChecks_KExceedsChecks(t *testing.T) {
	entries := buildClusterEntries() // 4 unique checks
	result := ClusterChecks(entries, 10)
	if len(result) != 4 {
		t.Fatalf("expected 4 clusters (capped), got %d", len(result))
	}
}

func TestClusterChecks_ClusterCount(t *testing.T) {
	entries := buildClusterEntries()
	result := ClusterChecks(entries, 2)
	if len(result) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(result))
	}
}

func TestClusterChecks_AllChecksAssigned(t *testing.T) {
	entries := buildClusterEntries()
	result := ClusterChecks(entries, 3)
	total := 0
	for _, c := range result {
		total += len(c.CheckNames)
	}
	if total != 4 {
		t.Fatalf("expected 4 checks assigned across clusters, got %d", total)
	}
}

func TestClusterChecks_MeanDurationPositive(t *testing.T) {
	entries := buildClusterEntries()
	result := ClusterChecks(entries, 2)
	for _, c := range result {
		if len(c.CheckNames) == 0 {
			continue
		}
		if c.MeanDuration <= 0 {
			t.Errorf("cluster %d has non-positive mean duration: %s", c.ClusterID, c.MeanDuration)
		}
	}
}

func TestClusterResult_String(t *testing.T) {
	c := ClusterResult{
		ClusterID:    1,
		CheckNames:   []string{"svc-a", "svc-b"},
		MeanDuration: 50 * time.Millisecond,
		StdDev:       5 * time.Millisecond,
	}
	s := c.String()
	for _, want := range []string{"Cluster 1", "svc-a", "svc-b", "50ms"} {
		if !containsSubstr(s, want) {
			t.Errorf("String() missing %q in: %s", want, s)
		}
	}
}

func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
