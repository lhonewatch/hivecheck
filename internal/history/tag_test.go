package history

import (
	"strings"
	"testing"
	"time"
)

func buildTaggedEntries() []TaggedEntry {
	now := time.Now()
	return []TaggedEntry{
		{CheckName: "api", Status: "ok", Duration: 10 * time.Millisecond, Timestamp: now, Tags: []string{"team:backend", "env:prod"}},
		{CheckName: "db", Status: "warn", Duration: 20 * time.Millisecond, Timestamp: now, Tags: []string{"team:backend", "env:staging"}},
		{CheckName: "cache", Status: "ok", Duration: 5 * time.Millisecond, Timestamp: now, Tags: []string{"team:infra", "env:prod"}},
		{CheckName: "worker", Status: "critical", Duration: 50 * time.Millisecond, Timestamp: now, Tags: []string{}},
	}
}

func TestHasTag_CaseInsensitive(t *testing.T) {
	e := TaggedEntry{Tags: []string{"Team:Backend"}}
	if !e.HasTag("team:backend") {
		t.Error("expected HasTag to match case-insensitively")
	}
	if e.HasTag("team:infra") {
		t.Error("expected HasTag to return false for absent tag")
	}
}

func TestFilterByTags_Empty(t *testing.T) {
	entries := buildTaggedEntries()
	f := TagFilter{}
	got := f.FilterByTags(entries)
	if len(got) != len(entries) {
		t.Errorf("expected all %d entries, got %d", len(entries), len(got))
	}
}

func TestFilterByTags_SingleTag(t *testing.T) {
	entries := buildTaggedEntries()
	f := TagFilter{Tags: []string{"env:prod"}}
	got := f.FilterByTags(entries)
	if len(got) != 2 {
		t.Errorf("expected 2 entries with env:prod, got %d", len(got))
	}
}

func TestFilterByTags_MultiTag(t *testing.T) {
	entries := buildTaggedEntries()
	f := TagFilter{Tags: []string{"team:backend", "env:prod"}}
	got := f.FilterByTags(entries)
	if len(got) != 1 {
		t.Errorf("expected 1 entry matching both tags, got %d", len(got))
	}
	if got[0].CheckName != "api" {
		t.Errorf("expected 'api', got %q", got[0].CheckName)
	}
}

func TestGroupByTag(t *testing.T) {
	entries := buildTaggedEntries()
	groups := GroupByTag(entries, "env:prod") // tag dimension doesn't split values here
	// worker has no tags → untagged
	if _, ok := groups["(untagged)"]; !ok {
		t.Error("expected (untagged) group for entry without matching tag")
	}
}

func TestFormatTagSummary_ContainsKeys(t *testing.T) {
	entries := buildTaggedEntries()
	groups := GroupByTag(entries, "team:backend")
	out := FormatTagSummary(groups)
	if !strings.Contains(out, "entries") {
		t.Errorf("expected 'entries' in output, got: %s", out)
	}
}
