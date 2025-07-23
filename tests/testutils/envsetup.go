package testutils

import "testing"

// SetTestEnv sets common environment variables required for tests.
func SetTestEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DISABLE_ENV_FILE", "true")
	t.Setenv("ADMIN_TELEGRAM_IDS", "1")
	t.Setenv("TELEGRAM_TOKEN", "t")
	t.Setenv("TRIAL_TRAFFIC_LIMIT", "1")
	t.Setenv("TRIAL_DAYS", "1")
	t.Setenv("PRICE_1", "1")
	t.Setenv("PRICE_3", "1")
	t.Setenv("PRICE_6", "1")
	t.Setenv("REMNAWAVE_URL", "http://example.com")
	t.Setenv("REMNAWAVE_TOKEN", "x")
	t.Setenv("DATABASE_URL", "db")
	t.Setenv("TRAFFIC_LIMIT", "1")
	t.Setenv("REFERRAL_DAYS", "0")
	t.Setenv("REFERRAL_BONUS", "0")
	t.Setenv("CRYPTO_PAY_ENABLED", "false")
	t.Setenv("TELEGRAM_STARS_ENABLED", "false")
}
