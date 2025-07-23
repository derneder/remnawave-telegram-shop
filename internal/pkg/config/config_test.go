package config_test

import (
	"testing"

	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/tests/testutils"
)

func TestInitConfigPrices(t *testing.T) {
	testutils.SetTestEnv(t)
	t.Setenv("TELEGRAM_TOKEN", "token")
	t.Setenv("PRICE_1", "10")
	t.Setenv("PRICE_3", "30")
	t.Setenv("PRICE_6", "50")
	t.Setenv("REMNAWAVE_URL", "http://example.com")
	t.Setenv("REMNAWAVE_TOKEN", "tok")
	t.Setenv("TRAFFIC_LIMIT", "100")
	t.Setenv("REFERRAL_BONUS", "150")

	config.InitConfig()

	if config.Price(3) != 30 {
		t.Fatalf("price3 expected 30, got %d", config.Price(3))
	}
	if config.Price(4) != 10 {
		t.Fatalf("default price expected 10, got %d", config.Price(4))
	}
}
