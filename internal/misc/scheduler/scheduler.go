package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/apolsh/yapr-gophkeeper/internal/logger"
)

var log = logger.LoggerOfComponent("scheduler")

type Scheduler struct {
	task         func(ctx context.Context) error
	errorHandler func(error)
	wg           sync.WaitGroup
}

func NewScheduler(task func(ctx context.Context) error, errorHandler func(error)) *Scheduler {
	return &Scheduler{task: task, errorHandler: errorHandler}
}

// RunWithInterval repeats the execution of tasks continuously.
// interval is set in seconds
func (s *Scheduler) RunWithInterval(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func(ctx context.Context, functionToRun func(ctx context.Context) error, errorHandler func(error)) {
		for {
			select {
			case <-ticker.C:
				s.wg.Add(1)
				err := functionToRun(ctx)
				if err != nil {
					errorHandler(err)
				}
				s.wg.Done()
			case <-ctx.Done():
				ticker.Stop()
			}
		}
	}(ctx, s.task, s.errorHandler)
}

// Close tries to wait for the current task to complete
func (s *Scheduler) Close() {
	waitWithTimeout(&s.wg, 5*time.Second)
}

func waitWithTimeout(wg *sync.WaitGroup, t time.Duration) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
	case <-time.After(t):
		log.Info("scheduler shutdown timeout exceeded")
	}
}
