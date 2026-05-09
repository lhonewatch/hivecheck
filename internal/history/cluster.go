package history

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// ClusterResult holds a group of checks that exhibit similar duration behaviour
// over the analysed window.
type ClusterResult struct {
	ClusterID   int
	CheckNames  []string
	MeanDuration time.Duration
	StdDev       time.Duration
}

// String returns a human-readable summary of the cluster.
func (c ClusterResult) String() string {
	names := strings.Join(c.CheckNames, ", ")
	return fmt.Sprintf("Cluster %d [mean=%s stddev=%s]: %s",
		c.ClusterID, c.MeanDuration.Round(time.Millisecond),
		c.StdDev.Round(time.Millisecond), names)
}

// ClusterChecks groups checks into clusters based on their average duration
// using a simple k-means-style approach (single pass, sorted bucketing).
// k controls the number of clusters; entries must contain duration data.
func ClusterChecks(entries []Entry, k int) []ClusterResult {
	if len(entries) == 0 || k <= 0 {
		return nil
	}

	// Aggregate mean duration per check name.
	type stat struct {
		name  string
		total time.Duration
		count int
	}
	agg := map[string]*stat{}
	for _, e := range entries {
		if e.Duration <= 0 {
			continue
		}
		if agg[e.CheckName] == nil {
			agg[e.CheckName] = &stat{name: e.CheckName}
		}
		agg[e.CheckName].total += e.Duration
		agg[e.CheckName].count++
	}

	type item struct {
		name string
		mean time.Duration
	}
	var items []item
	for _, s := range agg {
		if s.count == 0 {
			continue
		}
		items = append(items, item{name: s.name, mean: s.total / time.Duration(s.count)})
	}
	if len(items) == 0 {
		return nil
	}

	sort.Slice(items, func(i, j int) bool { return items[i].mean < items[j].mean })

	if k > len(items) {
		k = len(items)
	}

	// Divide sorted items into k equal-sized buckets.
	clusters := make([]ClusterResult, k)
	for i := range clusters {
		clusters[i].ClusterID = i + 1
	}
	for i, it := range items {
		bucket := int(float64(i) / float64(len(items)) * float64(k))
		if bucket >= k {
			bucket = k - 1
		}
		clusters[bucket].CheckNames = append(clusters[bucket].CheckNames, it.name)
	}

	// Compute mean and stddev per cluster.
	for ci := range clusters {
		var total time.Duration
		for _, name := range clusters[ci].CheckNames {
			total += agg[name].total / time.Duration(agg[name].count)
		}
		n := len(clusters[ci].CheckNames)
		if n == 0 {
			continue
		}
		clusters[ci].MeanDuration = total / time.Duration(n)

		var variance float64
		for _, name := range clusters[ci].CheckNames {
			d := float64(agg[name].total/time.Duration(agg[name].count) - clusters[ci].MeanDuration)
			variance += d * d
		}
		clusters[ci].StdDev = time.Duration(math.Sqrt(variance / float64(n)))
	}

	return clusters
}
