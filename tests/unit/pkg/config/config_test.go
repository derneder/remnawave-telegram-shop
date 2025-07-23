package config_test

import (
	"testing"

	"remnawave-tg-shop-bot/internal/pkg/config"
)

func TestInitConfigPrices(t *testing.T) {
	t.Setenv("DISABLE_ENV_FILE", "true")
	t.Setenv("ADMIN_TELEGRAM_IDS", "1")
	t.Setenv("TELEGRAM_TOKEN", "token")
	t.Setenv("TRIAL_TRAFFIC_LIMIT", "1")
	t.Setenv("TRIAL_DAYS", "1")
	t.Setenv("PRICE_1", "10")
	t.Setenv("PRICE_3", "30")
	t.Setenv("PRICE_6", "50")
	t.Setenv("REMNAWAVE_URL", "http://example.com")
	t.Setenv("REMNAWAVE_TOKEN", "tok")
	t.Setenv("DATABASE_URL", "db")
	t.Setenv("TRAFFIC_LIMIT", "100")
	t.Setenv("REFERRAL_DAYS", "0")
	t.Setenv("REFERRAL_BONUS", "150")
	t.Setenv("CRYPTO_PAY_ENABLED", "false")
	t.Setenv("TELEGRAM_STARS_ENABLED", "false")

	config.InitConfig()

	if config.Price(3) != 30 {
		t.Fatalf("price3 expected 30, got %d", config.Price(3))
	}
	if config.Price(4) != 10 {
		t.Fatalf("default price expected 10, got %d", config.Price(4))
	}
}
