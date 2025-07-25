package tests

import (
	"testing"

	"remnawave-tg-shop-bot/internal/pkg/config"
)

func TestInitConfigInvalidURL(t *testing.T) {
	SetTestEnv(t)
	t.Setenv("REMNAWAVE_URL", "://bad_url")
	if err := config.InitConfig(); err == nil {
		t.Fatal("expected error for invalid url")
	}
}

func TestInitConfigInvalidToken(t *testing.T) {
	SetTestEnv(t)
	t.Setenv("TELEGRAM_TOKEN", "bad token")
	if err := config.InitConfig(); err == nil {
		t.Fatal("expected error for invalid token")
	}
}
