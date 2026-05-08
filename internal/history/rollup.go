package history

import (
	"fmt"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// RollupPeriod defines the granularity of a rollup aggregation.
type RollupPeriod string

const (
	RollupHourly RollupPeriod = "hourly"
	RollupDaily  RollupPeriod = "daily"
)

// RollupBucket holds aggregated statistics for a single time bucket.
type RollupBucket struct {
	PeriodStart time.Time
	CheckName   string
	Total       int
	OKCount     int
	WarnCount   int
	CritCount   int
	ErrorCount  int
	AvgDuration time.Duration
}

// WorstStatus returns the worst status observed in this bucket.
func (b RollupBucket) WorstStatus() check.Status {
	switch {
	case b.ErrorCount > 0:
		return check.StatusError
	case b.CritCount > 0:
		return check.StatusCritical
	case b.WarnCount > 0:
		return check.StatusWarning
	default:
		return check.StatusOK
	}
}

// String returns a human-readable summary of the bucket.
func (b RollupBucket) String() string {
	return fmt.Sprintf("%s [%s] total=%d ok=%d warn=%d crit=%d err=%d avg=%s",
		b.CheckName,
		b.PeriodStart.Format(time.RFC3339),
		b.Total,
		b.OKCount,
		b.WarnCount,
		b.CritCount,
		b.ErrorCount,
		b.AvgDuration.Round(time.Millisecond),
	)
}

// RollupEntries aggregates a slice of StoreEntry values into buckets
// according to the requested period.
func RollupEntries(entries []StoreEntry, period RollupPeriod) []RollupBucket {
	type key struct {
		name   string
		bucket string
	}

	index := make(map[key]*RollupBucket)
	order := []key{}

	for _, e := range entries {
		var bucketKey string
		switch period {
		case RollupHourly:
			bucketKey = e.Timestamp.UTC().Format("2006-01-02T15")
		default:
			bucketKey = e.Timestamp.UTC().Format("2006-01-02")
		}

		k := key{name: e.CheckName, bucket: bucketKey}
		b, exists := index[k]
		if !exists {
			var t time.Time
			switch period {
			case RollupHourly:
				t, _ = time.Parse("2006-01-02T15", bucketKey)
			default:
				t, _ = time.Parse("2006-01-02", bucketKey)
			}
			b = &RollupBucket{PeriodStart: t.UTC(), CheckName: e.CheckName}
			index[k] = b
			order = append(order, k)
		}

		b.Total++
		b.AvgDuration += e.Duration
		switch e.Status {
		case check.StatusOK:
			b.OKCount++
		case check.StatusWarning:
			b.WarnCount++
		case check.StatusCritical:
			b.CritCount++
		case check.StatusError:
			b.ErrorCount++
		}
	}

	buckets := make([]RollupBucket, 0, len(order))
	for _, k := range order {
		b := index[k]
		if b.Total > 0 {
			b.AvgDuration = b.AvgDuration / time.Duration(b.Total)
		}
		buckets = append(buckets, *b)
	}
	return buckets
}
