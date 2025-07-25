package tribute

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubService struct{ calls int }

func (s *stubService) AddBalance(ctx context.Context, tg int64, amt int64) error {
	s.calls++
	return nil
}

func TestWebhookHandler(t *testing.T) {
	svc := &stubService{}
	h := NewHandler("key", svc)
	body := []byte(`{"event":"new_subscription","amount":10,"currency":"RUB","telegram_id":1}`)
	mac := hmac.New(sha256.New, []byte("key"))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r.Header.Set("trbt-signature", sig)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK || svc.calls != 1 {
		t.Fatalf("code %d calls %d", w.Code, svc.calls)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if svc.calls != 1 {
		t.Fatalf("duplicate call")
	}
	r = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r.Header.Set("trbt-signature", "bad")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden || svc.calls != 1 {
		t.Fatalf("expected 403 and single call")
	}
}
