package report

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// TextFormatter renders a Summary as human-readable plain text.
type TextFormatter struct{}

// Format writes a plain-text representation of s to w.
func (f *TextFormatter) Format(w io.Writer, s Summary) error {
	line := strings.Repeat("-", 60)
	fmt.Fprintln(w, line)
	fmt.Fprintf(w, "HiveCheck Report  %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintln(w, line)
	fmt.Fprintf(w, "Overall: %s\n", s.Overall)
	fmt.Fprintf(w, "Checks : %d  OK: %d  Warn: %d  Critical: %d  Error: %d\n",
		s.Total, s.Counts["ok"], s.Counts["warn"], s.Counts["critical"], s.Counts["error"])
	fmt.Fprintln(w, line)

	for _, r := range s.Results {
		duration := r.Duration.Round(time.Millisecond)
		fmt.Fprintf(w, "[%-8s] %-30s %s", r.Status, r.Name, duration)
		if r.Message != "" {
			fmt.Fprintf(w, "  %s", r.Message)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, line)
	return nil
}
