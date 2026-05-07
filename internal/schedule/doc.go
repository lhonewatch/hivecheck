// Package schedule provides a lightweight periodic scheduler for hivecheck
// health-check jobs.
//
// # Overview
//
// A [Scheduler] holds a set of [Job] values, each with its own polling
// interval and check function.  Calling [Scheduler.Start] launches one
// goroutine per job; each goroutine fires the check immediately and then
// repeats on the configured interval.  The scheduler runs until the supplied
// [context.Context] is cancelled, at which point all goroutines exit cleanly.
//
// # Usage
//
//	s := schedule.New(func(r check.Result) {
//		fmt.Println(r)
//	})
//	s.Add(schedule.Job{
//		Name:     "api-health",
//		Interval: 30 * time.Second,
//		Run:      myCheckFn,
//	})
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	s.Start(ctx)
package schedule
