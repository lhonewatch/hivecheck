package check

import (
	"context"
	"time"
)

// CheckFunc is a function that performs a health check and returns a raw numeric value.
type CheckFunc func(ctx context.Context) (value float64, message string, err error)

// Definition ties a named check to its thresholds and execution logic.
type Definition struct {
	Name      string
	Threshold Threshold
	Fn        CheckFunc
}

// Runner executes a set of check definitions with a shared timeout.
type Runner struct {
	Timeout     time.Duration
	Definitions []Definition
}

// NewRunner creates a Runner with the given timeout.
func NewRunner(timeout time.Duration, defs []Definition) *Runner {
	return &Runner{
		Timeout:     timeout,
		Definitions: defs,
	}
}

// Run executes all definitions concurrently and returns their results.
func (r *Runner) Run(ctx context.Context) []Result {
	ctx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()

	resultCh := make(chan Result, len(r.Definitions))

	for _, def := range r.Definitions {
		def := def
		go func() {
			start := time.Now()
			value, msg, err := def.Fn(ctx)
			dur := time.Since(start)

			res := Result{
				Name:      def.Name,
				Threshold: def.Threshold,
				Value:     value,
				Message:   msg,
				Duration:  dur,
				Err:       err,
			}
			if err != nil {
				res.Status = StatusUnknown
			} else {
				res.Status = def.Threshold.Evaluate(value)
			}
			resultCh <- res
		}()
	}

	results := make([]Result, 0, len(r.Definitions))
	for range r.Definitions {
		results = append(results, <-resultCh)
	}
	return results
}
