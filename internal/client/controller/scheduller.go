package controller

import (
	"context"
	"time"
)

func runPeriodically(ctx context.Context, functionToRun func(ctx context.Context) error, errorHandler func(error), periodInSeconds int64) {
	ticker := time.NewTicker(time.Duration(periodInSeconds) * time.Second)
	go func(ctx context.Context, functionToRun func(ctx context.Context) error, errorHandler func(error), periodInSeconds int64) {
		for {
			select {
			case <-ticker.C:
				err := functionToRun(ctx)
				if err != nil {
					errorHandler(err)
				}
			case <-ctx.Done():
				ticker.Stop()
			}
		}
	}(ctx, functionToRun, errorHandler, periodInSeconds)
}
