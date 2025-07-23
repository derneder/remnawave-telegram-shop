package tests

import (
	"context"
	"testing"

	"remnawave-tg-shop-bot/internal/app"
)

func TestInitDatabaseError(t *testing.T) {
	if _, err := app.InitDatabase(context.Background(), "bad://"); err == nil {
		t.Fatal("expected error")
	}
}
