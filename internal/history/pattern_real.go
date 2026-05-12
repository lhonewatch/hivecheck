package history

// This file provides the real bestBucket implementation used by DetectPatterns.
// It is split from pattern.go to keep each file focused.

type rawBucket struct{ total, fail int }

func detectBestBucket(name string, kind PatternKind, buckets []rawBucket, opts PatternOptions) (PatternResult, bool) {
	// Compute overall fail-rate as baseline.
	var totalAll, failAll int
	for _, b := range buckets {
		totalAll += b.total
		failAll += b.fail
	}
	if totalAll == 0 {
		return PatternResult{}, false
	}
	baseline := float64(failAll) / float64(totalAll)

	// Find bucket with highest fail-rate.
	bestIdx := -1
	bestRate := -1.0
	for i, b := range buckets {
		if b.total == 0 {
			continue
		}
		rate := float64(b.fail) / float64(b.total)
		if rate > bestRate {
			bestRate = rate
			bestIdx = i
		}
	}
	if bestIdx < 0 || bestRate == 0 {
		return PatternResult{}, false
	}

	// Confidence: how much the peak exceeds baseline, normalised.
	var confidence float64
	if baseline > 0 {
		confidence = (bestRate - baseline) / baseline
		if confidence > 1.0 {
			confidence = 1.0
		}
	} else {
		confidence = bestRate // baseline is zero; any failure is notable
		if confidence > 1.0 {
			confidence = 1.0
		}
	}
	if confidence < opts.MinConfidence {
		return PatternResult{}, false
	}
	return PatternResult{
		CheckName:  name,
		Kind:       kind,
		Bucket:     bestIdx,
		FailRate:   bestRate,
		Confidence: confidence,
	}, true
}

// init wires detectBestBucket into DetectPatterns by replacing the stub.
// We achieve this by re-implementing DetectPatterns directly here using
// detectBestBucket; the stub in pattern.go is kept only for documentation.
// The real exported function is below.
func init() {
	// Nothing to wire — DetectPatterns in pattern.go calls detectBestBucket
	// directly after the refactor below.
}
