package history

import (
	"fmt"
	"math"
	"time"
)

// ForecastResult holds a simple linear-regression forecast for a single check.
type ForecastResult struct {
	CheckName    string
	Slope        float64       // ms per second trend in latency
	Predicted    float64       // predicted avg duration (ms) at Horizon
	Horizon      time.Time     // point in time the prediction targets
	Confidence   string        // "low" | "medium" | "high" based on sample count
	Degrading    bool          // true when slope indicates worsening latency
}

// String returns a human-readable summary of the forecast.
func (f ForecastResult) String() string {
	direction := "stable"
	if f.Degrading {
		direction = "degrading"
	}
	return fmt.Sprintf("check=%s horizon=%s predicted=%.1fms slope=%+.4f/s confidence=%s trend=%s",
		f.CheckName,
		f.Horizon.Format(time.RFC3339),
		f.Predicted,
		f.Slope,
		f.Confidence,
		direction,
	)
}

// ForecastChecks performs a per-check linear regression over the supplied
// entries and returns a forecast at now+horizon for each check that has
// enough data points (minimum 3).
func ForecastChecks(entries []Entry, horizon time.Duration) []ForecastResult {
	if len(entries) == 0 || horizon <= 0 {
		return nil
	}

	type point struct {
		t float64 // unix seconds
		v float64 // duration ms
	}

	byCheck := make(map[string][]point)
	for _, e := range entries {
		byCheck[e.CheckName] = append(byCheck[e.CheckName], point{
			t: float64(e.Timestamp.Unix()),
			v: float64(e.DurationMS),
		})
	}

	now := time.Now()
	target := float64(now.Add(horizon).Unix())

	var results []ForecastResult
	for name, pts := range byCheck {
		if len(pts) < 3 {
			continue
		}
		slope, intercept := linearRegression(pts)
		predicted := slope*target + intercept
		if predicted < 0 {
			predicted = 0
		}
		results = append(results, ForecastResult{
			CheckName:  name,
			Slope:      slope,
			Predicted:  math.Round(predicted*100) / 100,
			Horizon:    now.Add(horizon),
			Confidence: confidenceLabel(len(pts)),
			Degrading:  slope > 0,
		})
	}
	return results
}

func linearRegression(pts []struct{ t, v float64 }) (slope, intercept float64) {
	n := float64(len(pts))
	var sumT, sumV, sumTT, sumTV float64
	for _, p := range pts {
		sumT += p.t
		sumV += p.v
		sumTT += p.t * p.t
		sumTV += p.t * p.v
	}
	denom := n*sumTT - sumT*sumT
	if denom == 0 {
		return 0, sumV / n
	}
	slope = (n*sumTV - sumT*sumV) / denom
	intercept = (sumV - slope*sumT) / n
	return
}

func confidenceLabel(n int) string {
	switch {
	case n >= 20:
		return "high"
	case n >= 7:
		return "medium"
	default:
		return "low"
	}
}
