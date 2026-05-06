// Package report provides formatting and summarization of check results.
package report

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// Summary holds aggregated results from a check run.
type Summary struct {
	Timestamp time.Time
	Results   []check.Result
	Duration  time.Duration
}

// StatusCounts returns a map of status -> count across all results.
func (s *Summary) StatusCounts() map[check.Status]int {
	counts := make(map[check.Status]int)
	for _, r := range s.Results {
		counts[r.Status]++
	}
	return counts
}

// Overall returns the worst status observed across all results.
func (s *Summary) Overall() check.Status {
	worst := check.StatusOK
	for _, r := range s.Results {
		if r.Status > worst {
			worst = r.Status
		}
	}
	return worst
}

// Formatter writes a Summary to an io.Writer.
type Formatter interface {
	Format(w io.Writer, s *Summary) error
}

// TextFormatter renders results as human-readable text.
type TextFormatter struct{}

// Format writes a plain-text report to w.
func (f *TextFormatter) Format(w io.Writer, s *Summary) error {
	fmt.Fprintf(w, "HiveCheck Report — %s\n", s.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(w, "Duration : %s\n", s.Duration.Round(time.Millisecond))
	fmt.Fprintf(w, "Overall  : %s\n", s.Overall())
	fmt.Fprintln(w, strings.Repeat("-", 48))
	for _, r := range s.Results {
		line := fmt.Sprintf("  [%s] %s", r.Status, r.Name)
		if r.Message != "" {
			line += " — " + r.Message
		}
		if r.Err != nil {
			line += fmt.Sprintf(" (error: %v)", r.Err)
		}
		fmt.Fprintln(w, line)
	}
	counts := s.StatusCounts()
	fmt.Fprintf(w, strings.Repeat("-", 48)+"\n")
	fmt.Fprintf(w, "OK=%d  WARN=%d  CRIT=%d  ERR=%d\n",
		counts[check.StatusOK],
		counts[check.StatusWarn],
		counts[check.StatusCritical],
		counts[check.StatusError],
	)
	return nil
}
