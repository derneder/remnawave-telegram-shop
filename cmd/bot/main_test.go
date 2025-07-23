package main

import "testing"

func TestMainSmoke(t *testing.T) {
	t.Setenv("DISABLE_ENV_FILE", "true")
	t.Setenv("ADMIN_TELEGRAM_IDS", "1")
	t.Setenv("TELEGRAM_TOKEN", "t")
	t.Setenv("TRIAL_TRAFFIC_LIMIT", "1")
	t.Setenv("TRIAL_DAYS", "1")
	t.Setenv("PRICE_1", "10")
	t.Setenv("PRICE_3", "30")
	t.Setenv("PRICE_6", "50")
	t.Setenv("REMNAWAVE_URL", "http://x")
	t.Setenv("REMNAWAVE_TOKEN", "x")
	t.Setenv("DATABASE_URL", "bad://")
	t.Setenv("TRAFFIC_LIMIT", "100")
	t.Setenv("REFERRAL_DAYS", "0")
	t.Setenv("REFERRAL_BONUS", "0")
	t.Setenv("CRYPTO_PAY_ENABLED", "false")
	t.Setenv("TELEGRAM_STARS_ENABLED", "false")
	main()
}
