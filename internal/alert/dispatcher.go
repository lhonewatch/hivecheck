package alert

import (
	"fmt"
	"log"

	"github.com/example/hivecheck/internal/check"
)

// Hook is the interface that alert backends must implement.
type Hook interface {
	Send(result check.Result) error
}

// Dispatcher fans out check results to one or more hooks based on status.
type Dispatcher struct {
	hooks      []Hook
	minStatus  check.Status
	logFailure bool
}

// DispatcherOption configures a Dispatcher.
type DispatcherOption func(*Dispatcher)

// WithMinStatus sets the minimum status level that triggers dispatch.
func WithMinStatus(s check.Status) DispatcherOption {
	return func(d *Dispatcher) { d.minStatus = s }
}

// WithLogFailure controls whether hook errors are logged instead of returned.
func WithLogFailure(v bool) DispatcherOption {
	return func(d *Dispatcher) { d.logFailure = v }
}

// NewDispatcher creates a Dispatcher with the given hooks and options.
func NewDispatcher(hooks []Hook, opts ...DispatcherOption) *Dispatcher {
	d := &Dispatcher{
		hooks:     hooks,
		minStatus: check.StatusWarn,
	}
	for _, o := range opts {
		o(d)
	}
	return d
}

// Dispatch sends the result to all registered hooks if the result status
// meets or exceeds the configured minimum status threshold.
func (d *Dispatcher) Dispatch(result check.Result) error {
	if result.Status < d.minStatus {
		return nil
	}

	var firstErr error
	for _, h := range d.hooks {
		if err := h.Send(result); err != nil {
			if d.logFailure {
				log.Printf("alert hook error for check %q: %v", result.Name, err)
				continue
			}
			if firstErr == nil {
				firstErr = fmt.Errorf("hook send failed for check %q: %w", result.Name, err)
			}
		}
	}
	return firstErr
}
