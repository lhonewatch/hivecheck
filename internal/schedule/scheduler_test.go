package schedule_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
	"github.com/example/hivecheck/internal/schedule"
)

func makeJob(name string, interval time.Duration, calls *int32) schedule.Job {
	return schedule.Job{
		Name:     name,
		Interval: interval,
		Run: func(_ context.Context) (check.Result, error) {
			atomic.AddInt32(calls, 1)
			return check.Result{Name: name, Status: check.StatusOK}, nil
		},
	}
}

func TestScheduler_RunsImmediately(t *testing.T) {
	var calls int32
	var results []check.Result
	var mu sync.Mutex

	s := schedule.New(func(r check.Result) {
		mu.Lock()
		results = append(results, r)
		mu.Unlock()
	})
	s.Add(makeJob("immediate", 10*time.Second, &calls))

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	s.Start(ctx)

	if atomic.LoadInt32(&calls) < 1 {
		t.Fatal("expected at least one immediate execution")
	}
	mu.Lock()
	defer mu.Unlock()
	if len(results) < 1 {
		t.Fatal("expected at least one result delivered to callback")
	}
	if results[0].Name != "immediate" {
		t.Errorf("unexpected result name: %s", results[0].Name)
	}
}

func TestScheduler_MultipleJobs(t *testing.T) {
	var callsA, callsB int32

	s := schedule.New(nil)
	s.Add(makeJob("a", 50*time.Millisecond, &callsA))
	s.Add(makeJob("b", 50*time.Millisecond, &callsB))

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Millisecond)
	defer cancel()
	s.Start(ctx)

	if atomic.LoadInt32(&callsA) < 2 {
		t.Errorf("job a: expected >=2 calls, got %d", callsA)
	}
	if atomic.LoadInt32(&callsB) < 2 {
		t.Errorf("job b: expected >=2 calls, got %d", callsB)
	}
}

func TestScheduler_CancelStopsJobs(t *testing.T) {
	var calls int32
	s := schedule.New(nil)
	s.Add(makeJob("cancel-test", 20*time.Millisecond, &calls))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	s.Start(ctx)

	// Only the immediate call should have happened (if any).
	if atomic.LoadInt32(&calls) > 1 {
		t.Errorf("expected at most 1 call after immediate cancel, got %d", calls)
	}
}
