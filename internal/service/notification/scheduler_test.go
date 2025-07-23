package notification

import (
	"context"
	"testing"

	"github.com/robfig/cron/v3"
)

type mockNotifier struct{ called bool }

func (m *mockNotifier) SendSubscriptionNotifications(ctx context.Context) error {
	m.called = true
	return nil
}

func TestRegisterSubscriptionCron(t *testing.T) {
	c := cron.New()
	m := &mockNotifier{}
	if err := RegisterSubscriptionCron(c, m); err != nil {
		t.Fatalf("register cron: %v", err)
	}
	entries := c.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	entries[0].Job.Run()
	if !m.called {
		t.Fatalf("expected SendSubscriptionNotifications to be called")
	}
}
