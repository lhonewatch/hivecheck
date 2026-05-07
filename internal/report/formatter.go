package report

import (
	"fmt"
	"io"
)

// Formatter is the interface implemented by report formatters.
type Formatter interface {
	Format(w io.Writer, s Summary) error
}

// NewFormatter returns a Formatter for the given format name.
// Supported values: "text", "json". Defaults to "text" for unknown values.
func NewFormatter(format string) (Formatter, error) {
	switch format {
	case "json":
		return &JSONFormatter{}, nil
	case "text", "":
		return &TextFormatter{}, nil
	default:
		return nil, fmt.Errorf("report: unknown format %q", format)
	}
}
