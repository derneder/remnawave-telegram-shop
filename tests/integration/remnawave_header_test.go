//go:build integration

package remnawave

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/tests/testutils"
)

func TestHeaderTransport(t *testing.T) {
	testutils.SetTestEnv(t)
	t.Setenv("REMNAWAVE_URL", "")
	t.Setenv("REFERRAL_BONUS", "0")
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
