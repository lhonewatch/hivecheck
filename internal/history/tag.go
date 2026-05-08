package history

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// TagFilter selects history entries that carry all of the specified tags.
type TagFilter struct {
	Tags []string
}

// TaggedEntry wraps a StoreEntry with an optional set of string tags.
type TaggedEntry struct {
	CheckName string
	Status    string
	Duration  time.Duration
	Timestamp time.Time
	Tags      []string
}

// HasTag reports whether the entry carries the given tag (case-insensitive).
func (e TaggedEntry) HasTag(tag string) bool {
	tag = strings.ToLower(tag)
	for _, t := range e.Tags {
		if strings.ToLower(t) == tag {
			return true
		}
	}
	return false
}

// FilterByTags returns only those entries that carry every tag in f.Tags.
func (f TagFilter) FilterByTags(entries []TaggedEntry) []TaggedEntry {
	if len(f.Tags) == 0 {
		return entries
	}
	out := make([]TaggedEntry, 0, len(entries))
	for _, e := range entries {
		if f.matchAll(e) {
			out = append(out, e)
		}
	}
	return out
}

func (f TagFilter) matchAll(e TaggedEntry) bool {
	for _, t := range f.Tags {
		if !e.HasTag(t) {
			return false
		}
	}
	return true
}

// GroupByTag groups entries by a single tag dimension, returning a map of
// tag value → entries. Entries without the tag are placed under "(untagged)".
func GroupByTag(entries []TaggedEntry, tag string) map[string][]TaggedEntry {
	result := make(map[string][]TaggedEntry)
	tag = strings.ToLower(tag)
	for _, e := range entries {
		key := "(untagged)"
		for _, t := range e.Tags {
			if strings.ToLower(t) == tag {
				key = t
				break
			}
		}
		result[key] = append(result[key], e)
	}
	return result
}

// FormatTagSummary returns a human-readable summary of entries per tag group.
func FormatTagSummary(groups map[string][]TaggedEntry) string {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("[%s] %d entries\n", k, len(groups[k])))
	}
	return strings.TrimRight(sb.String(), "\n")
}
