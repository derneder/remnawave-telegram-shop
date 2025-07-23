package app_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

func TestCronRunsJob(t *testing.T) {
	c := cron.New(cron.WithSeconds())

	var ran atomic.Bool
	ctx, cancel := context.WithCancel(context.Background())

	_, err := c.AddFunc("@every 1s", func() {
		ran.Store(true)
		cancel()
	})
	if err != nil {
		t.Fatalf("add func: %v", err)
	}

	c.Start()
	<-ctx.Done()
	stopCtx := c.Stop()
	select {
	case <-stopCtx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("cron did not stop")
	}

	if !ran.Load() {
		t.Fatal("job did not run")
	}
}
