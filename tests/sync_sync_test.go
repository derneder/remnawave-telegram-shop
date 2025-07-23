package tests

import (
	"testing"

	syncsvc "remnawave-tg-shop-bot/internal/service/sync"
)

func TestNewSyncService(t *testing.T) {
	s := syncsvc.NewSyncService(nil, nil)
	if s == nil {
		t.Fatal("nil service")
	}
}
