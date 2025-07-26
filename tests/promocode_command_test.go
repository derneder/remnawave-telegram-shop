package tests

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	handlerpkg "remnawave-tg-shop-bot/internal/adapter/telegram/handler"
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

type stubHTTPPromoCmd struct{ body string }

func (h *stubHTTPPromoCmd) Do(r *http.Request) (*http.Response, error) {
	b, _ := io.ReadAll(r.Body)
	h.body = string(b)
	resp := &http.Response{StatusCode: http.StatusOK}
	resp.Body = io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`))
	return resp, nil
}

func TestPromocodeCommandHandler_NoArgs(t *testing.T) {
	SetTestEnv(t)
	tm := translation.GetInstance()
	_ = tm.InitDefaultTranslations()
	httpc := &stubHTTPPromoCmd{}
	b, _ := bot.New("t", bot.WithHTTPClient(time.Second, httpc), bot.WithSkipGetMe())
	h := handlerpkg.NewHandler(nil, nil, tm, &StubCustomerRepo{}, nil, nil, nil, nil, nil, nil)

	upd := &models.Update{Message: &models.Message{Chat: models.Chat{ID: 1}, From: &models.User{ID: 1, LanguageCode: "ru"}, Text: "/promocode"}}
	h.PromocodeCommandHandler(context.Background(), b, upd)

	if !strings.Contains(httpc.body, tm.GetText("ru", "promo.activate.prompt")) {
		t.Fatalf("unexpected body: %s", httpc.body)
	}
	if !h.IsAwaitingPromo(1) {
		t.Fatal("state not set")
	}
}

func TestPromoCommandHandler_NoArgs(t *testing.T) {
	SetTestEnv(t)
	tm := translation.GetInstance()
	_ = tm.InitDefaultTranslations()
	httpc := &stubHTTPPromoCmd{}
	b, _ := bot.New("t", bot.WithHTTPClient(time.Second, httpc), bot.WithSkipGetMe())
	h := handlerpkg.NewHandler(nil, nil, tm, &StubCustomerRepo{}, nil, nil, nil, nil, nil, nil)

	upd := &models.Update{Message: &models.Message{Chat: models.Chat{ID: 1}, From: &models.User{ID: 1, LanguageCode: "ru"}, Text: "/promo"}}
	h.PromoCommandHandler(context.Background(), b, upd)

	if !strings.Contains(httpc.body, tm.GetText("ru", "promo.activate.prompt")) {
		t.Fatalf("unexpected body: %s", httpc.body)
	}
	if !h.IsAwaitingPromo(1) {
		t.Fatal("state not set")
	}
}
