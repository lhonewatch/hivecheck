package check

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunnerRun_OK(t *testing.T) {
	defs := []Definition{
		{
			Name:      "latency",
			Threshold: Threshold{Warn: 200, Critical: 500},
			Fn: func(ctx context.Context) (float64, string, error) {
				return 120, "latency within bounds", nil
			},
		},
	}

	runner := NewRunner(5*time.Second, defs)
	results := runner.Run(context.Background())

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusOK {
		t.Errorf("expected OK, got %s", results[0].Status)
	}
}

func TestRunnerRun_Warn(t *testing.T) {
	defs := []Definition{
		{
			Name:      "cpu",
			Threshold: Threshold{Warn: 70, Critical: 90},
			Fn: func(ctx context.Context) (float64, string, error) {
				return 75, "cpu usage elevated", nil
			},
		},
	}

	runner := NewRunner(5*time.Second, defs)
	results := runner.Run(context.Background())

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusWarn {
		t.Errorf("expected WARN, got %s", results[0].Status)
	}
}

func TestRunnerRun_Critical(t *testing.T) {
	defs := []Definition{
		{
			Name:      "error_rate",
			Threshold: Threshold{Warn: 5, Critical: 10},
			Fn: func(ctx context.Context) (float64, string, error) {
				return 15, "high error rate", nil
			},
		},
	}

	runner := NewRunner(5*time.Second, defs)
	results := runner.Run(context.Background())

	if results[0].Status != StatusCritical {
		t.Errorf("expected CRITICAL, got %s", results[0].Status)
	}
}

func TestRunnerRun_Error(t *testing.T) {
	defs := []Definition{
		{
			Name:      "disk",
			Threshold: Threshold{Warn: 70, Critical: 90},
			Fn: func(ctx context.Context) (float64, string, error) {
				return 0, "", errors.New("disk unreachable")
			},
		},
	}

	runner := NewRunner(5*time.Second, defs)
	results := runner.Run(context.Background())

	if results[0].Status != StatusUnknown {
		t.Errorf("expected UNKNOWN, got %s", results[0].Status)
	}
	if results[0].Err == nil {
		t.Error("expected non-nil error")
	}
}

func TestRunnerRun_Concurrent(t *testing.T) {
	defs := make([]Definition, 10)
	for i := range defs {
		defs[i] = Definition{
			Name:      "check",
			Threshold: Threshold{Warn: 50, Critical: 90},
			Fn: func(ctx context.Context) (float64, string, error) {
				time.Sleep(10 * time.Millisecond)
				return 30, "ok", nil
			},
		}
	}

	runner := NewRunner(5*time.Second, defs)
	start := time.Now()
	results := runner.Run(context.Background())
	elapsed := time.Since(start)

	if len(results) != 10 {
		t.Fatalf("expected 10 results, got %d", len(results))
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("checks took too long: %s (expected concurrent execution)", elapsed)
	}
}
