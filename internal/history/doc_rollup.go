// Package history provides facilities for persisting, querying, and
// analysing historical check results produced by hivecheck.
//
// # Rollup
//
// The rollup sub-feature aggregates individual StoreEntry records into
// time-bucketed summaries (RollupBucket) at hourly or daily granularity.
// This is useful for dashboards and long-term trend analysis where
// per-run detail is not required.
//
// Usage:
//
//	entries, _ := store.Load("my-check", 0)
//	buckets := history.RollupEntries(entries, history.RollupDaily)
//	for _, b := range buckets {
//		fmt.Println(b)
//	}
//
// # Rollup CSV Export
//
// ExportRollupCSV serialises a slice of RollupBucket values to an
// io.Writer in CSV format, suitable for import into spreadsheet tools
// or time-series databases.
package history
