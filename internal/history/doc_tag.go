// Package history provides utilities for recording, querying, and analysing
// the historical results of hivecheck health checks.
//
// # Tag filtering
//
// The tag sub-feature allows callers to attach arbitrary string labels (tags)
// to history entries and later filter or group those entries by tag.
//
// Tags follow a "key:value" convention by recommendation, but any non-empty
// string is accepted.  Matching is always case-insensitive.
//
// Example usage:
//
//	filter := history.TagFilter{Tags: []string{"env:prod", "team:backend"}}
//	matched := filter.FilterByTags(allEntries)
//
//	groups := history.GroupByTag(allEntries, "env:prod")
//	fmt.Println(history.FormatTagSummary(groups))
package history
