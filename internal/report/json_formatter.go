package report

import (
	"encoding/json"
	"io"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// jsonResult is the wire representation of a single check result.
type jsonResult struct {
	Name    string      `json:"name"`
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Value   float64     `json:"value"`
	Err     string      `json:"error,omitempty"`
}

// jsonReport is the top-level JSON document.
type jsonReport struct {
	Timestamp time.Time    `json:"timestamp"`
	DurationMs int64       `json:"duration_ms"`
	Overall   string       `json:"overall"`
	Results   []jsonResult `json:"results"`
}

// JSONFormatter renders results as a JSON document.
type JSONFormatter struct {
	Indent bool
}

// Format writes a JSON-encoded report to w.
func (f *JSONFormatter) Format(w io.Writer, s *Summary) error {
	rpt := jsonReport{
		Timestamp:  s.Timestamp,
		DurationMs: s.Duration.Milliseconds(),
		Overall:    s.Overall().String(),
		Results:    make([]jsonResult, 0, len(s.Results)),
	}
	for _, r := range s.Results {
		jr := jsonResult{
			Name:    r.Name,
			Status:  r.Status.String(),
			Message: r.Message,
			Value:   r.Value,
		}
		if r.Err != nil {
			jr.Err = r.Err.Error()
		}
		rpt.Results = append(rpt.Results, jr)
	}
	enc := json.NewEncoder(w)
	if f.Indent {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(rpt)
}

// Ensure JSONFormatter satisfies Formatter at compile time.
var _ Formatter = (*JSONFormatter)(nil)

// toStatus maps a check.Result field for reuse.
func toStatus(r check.Result) string { return r.Status.String() }
