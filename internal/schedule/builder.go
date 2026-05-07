package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/example/hivecheck/internal/check"
	"github.com/example/hivecheck/internal/config"
)

// FromConfig builds a slice of Jobs from a loaded configuration, wiring each
// configured check to the provided Runner.
func FromConfig(cfg *config.Config, runner *check.Runner) ([]Job, error) {
	if len(cfg.Checks) == 0 {
		return nil, fmt.Errorf("schedule: no checks defined in config")
	}

	interval := time.Duration(cfg.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 60 * time.Second
	}

	jobs := make([]Job, 0, len(cfg.Checks))
	for _, chk := range cfg.Checks {
		chk := chk // capture loop variable
		jobs = append(jobs, Job{
			Name:     chk.Name,
			Interval: interval,
			Run: func(ctx context.Context) (check.Result, error) {
				return runner.Run(ctx, chk)
			},
		})
	}
	return jobs, nil
}
