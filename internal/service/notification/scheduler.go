package notification

import (
	"context"
	"github.com/robfig/cron/v3"
	"log/slog"
)

type subscriptionNotifier interface {
	SendSubscriptionNotifications(ctx context.Context) error
}

func RegisterSubscriptionCron(c *cron.Cron, svc subscriptionNotifier) error {
	_, err := c.AddFunc("@daily", func() {
		if err := svc.SendSubscriptionNotifications(context.Background()); err != nil {
			slog.Error("send subscription notifications", "err", err)
		}
	})
	return err
}
