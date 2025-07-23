//go:build integration

package remnawave

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"remnawave-tg-shop-bot/internal/pkg/config"
)

func TestHeaderTransport(t *testing.T) {
	t.Setenv("DISABLE_ENV_FILE", "true")
	t.Setenv("ADMIN_TELEGRAM_IDS", "1")
	t.Setenv("TELEGRAM_TOKEN", "t")
	t.Setenv("TRIAL_TRAFFIC_LIMIT", "1")
	t.Setenv("TRIAL_DAYS", "1")
	t.Setenv("PRICE_1", "1")
	t.Setenv("PRICE_3", "1")
	t.Setenv("PRICE_6", "1")
	t.Setenv("REMNAWAVE_URL", "")
	t.Setenv("REMNAWAVE_TOKEN", "x")
	t.Setenv("DATABASE_URL", "db")
	t.Setenv("TRAFFIC_LIMIT", "1")
	t.Setenv("REFERRAL_DAYS", "0")
	t.Setenv("REFERRAL_BONUS", "0")
	t.Setenv("CRYPTO_PAY_ENABLED", "false")
	t.Setenv("TELEGRAM_STARS_ENABLED", "false")
	t.Setenv("X_API_KEY", "key")

	config.InitConfig()

	var gotKey, gotForward string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Api-Key")
		gotForward = r.Header.Get("x-forwarded-for")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"users":[],"total":0}}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "token", "local")
	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("ping: %v", err)
	}

	if gotKey != "key" {
		t.Errorf("expected header key, got %q", gotKey)
	}
	if gotForward == "" {
		t.Errorf("x-forwarded-for not set")
	}
}
