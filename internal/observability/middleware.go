package observability

import (
	"context"
	"time"
)

type HandlerFunc func(ctx context.Context) error

func Measure(name string, next HandlerFunc) HandlerFunc {
	return func(ctx context.Context) error {
		start := time.Now()
		err := next(ctx)
		RequestDuration.WithLabelValues(name).Observe(time.Since(start).Seconds())
		return err
	}
}
