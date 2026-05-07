// Package schedule provides periodic execution of health checks.
package schedule

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// Job pairs a named interval with a check function.
type Job struct {
	Name     string
	Interval time.Duration
	Run      func(ctx context.Context) (check.Result, error)
}

// Scheduler runs Jobs on their configured intervals until the context is
// cancelled.
type Scheduler struct {
	jobs    []Job
	onResult func(check.Result)
	mu      sync.Mutex
}

// New returns a Scheduler that calls onResult for every completed check.
func New(onResult func(check.Result)) *Scheduler {
	return &Scheduler{onResult: onResult}
}

// Add registers a Job with the scheduler.
func (s *Scheduler) Add(j Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = append(s.jobs, j)
}

// Start launches a goroutine for each job and blocks until ctx is done.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	jobs := make([]Job, len(s.jobs))
	copy(jobs, s.jobs)
	s.mu.Unlock()

	var wg sync.WaitGroup
	for _, j := range jobs {
		wg.Add(1)
		go func(job Job) {
			defer wg.Done()
			s.runLoop(ctx, job)
		}(j)
	}
	wg.Wait()
}

func (s *Scheduler) runLoop(ctx context.Context, job Job) {
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	// Run once immediately before waiting for the first tick.
	s.execute(ctx, job)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.execute(ctx, job)
		}
	}
}

func (s *Scheduler) execute(ctx context.Context, job Job) {
	result, err := job.Run(ctx)
	if err != nil {
		log.Printf("scheduler: job %q error: %v", job.Name, err)
		return
	}
	if s.onResult != nil {
		s.onResult(result)
	}
}
