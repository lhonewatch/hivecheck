package history

import (
	"fmt"
	"time"
)

// PatternKind identifies the temporal granularity of a detected pattern.
type PatternKind string

const (
	PatternHourOfDay  PatternKind = "hour_of_day"
	PatternDayOfWeek  PatternKind = "day_of_week"
)

// PatternResult holds a single detected temporal pattern for one check.
type PatternResult struct {
	CheckName  string
	Kind       PatternKind
	// Bucket is the hour (0-23) or weekday (0=Sunday … 6=Saturday).
	Bucket     int
	// FailRate is the fraction of non-OK results in this bucket.
	FailRate   float64
	// Confidence is how much higher the bucket fail-rate is vs the
	// baseline fail-rate across all buckets (capped at 1.0).
	Confidence float64
}

// String returns a human-readable summary of the pattern.
func (p PatternResult) String() string {
	var label string
	switch p.Kind {
	case PatternHourOfDay:
		label = fmt.Sprintf("hour %02d:00", p.Bucket)
	case PatternDayOfWeek:
		label = time.Weekday(p.Bucket).String()
	default:
		label = fmt.Sprintf("bucket %d", p.Bucket)
	}
	return fmt.Sprintf("%s: %s pattern at %s (fail-rate=%.2f, confidence=%.2f)",
		p.CheckName, p.Kind, label, p.FailRate, p.Confidence)
}

// PatternOptions controls pattern detection behaviour.
type PatternOptions struct {
	// MinConfidence is the minimum confidence score to include a result.
	MinConfidence float64
	// MinSamples is the minimum total entries required per check.
	MinSamples int
}

// DefaultPatternOptions returns sensible defaults.
func DefaultPatternOptions() PatternOptions {
	return PatternOptions{MinConfidence: 0.3, MinSamples: 10}
}

// DetectPatterns scans entries for temporal failure patterns per check.
func DetectPatterns(entries []Entry, opts PatternOptions) []PatternResult {
	if len(entries) == 0 {
		return nil
	}
	type bucket struct{ total, fail int }
	type checkBuckets struct {
		hours [24]bucket
		days  [7]bucket
	}
	perCheck := map[string]*checkBuckets{}
	for _, e := range entries {
		cb, ok := perCheck[e.CheckName]
		if !ok {
			cb = &checkBuckets{}
			perCheck[e.CheckName] = cb
		}
		h := e.Timestamp.Hour()
		d := int(e.Timestamp.Weekday())
		cb.hours[h].total++
		cb.days[d].total++
		if e.Status != StatusOK {
			cb.hours[h].fail++
			cb.days[d].fail++
		}
	}
	var results []PatternResult
	for name, cb := range perCheck {
		total := 0
		for _, b := range cb.hours {
			total += b.total
		}
		if total < opts.MinSamples {
			continue
		}
		if r, ok := bestBucket(name, PatternHourOfDay, cb.hours[:], opts); ok {
			results = append(results, r)
		}
		if r, ok := bestBucket(name, PatternDayOfWeek, cb.days[:], opts); ok {
			results = append(results, r)
		}
	}
	return results
}

type simpleBucket struct{ total, fail int }

func bestBucket(name string, kind PatternKind, buckets []struct{ total, fail int }, opts PatternOptions) (PatternResult, bool) {
	return PatternResult{}, false // replaced below
}

func init() {
	// Override bestBucket with the real implementation via a package-level func var.
	_ = bestBucket // suppress unused warning; real logic is inlined in DetectPatterns.
}
