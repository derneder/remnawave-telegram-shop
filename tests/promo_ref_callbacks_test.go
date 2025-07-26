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
	"remnawave-tg-shop-bot/internal/pkg/contextkey"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	uimenu "remnawave-tg-shop-bot/internal/ui/menu"
)

type stubHTTPPromo struct{ body string }

func (h *stubHTTPPromo) Do(r *http.Request) (*http.Response, error) {
	b, _ := io.ReadAll(r.Body)
	h.body = string(b)
	resp := &http.Response{StatusCode: http.StatusOK}
	resp.Body = io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`))
	return resp, nil
}

func TestPromoActivateCallback(t *testing.T) {
	SetTestEnv(t)
	tm := translation.GetInstance()
	_ = tm.InitDefaultTranslations()
	httpc := &stubHTTPPromo{}
	b, _ := bot.New("t", bot.WithHTTPClient(time.Second, httpc), bot.WithSkipGetMe())
	h := handlerpkg.NewHandler(nil, nil, tm, &StubCustomerRepo{}, nil, nil, nil, nil, nil, nil)

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{ID: "1", Data: uimenu.CallbackPromoUserActivate, From: models.User{ID: 1, LanguageCode: "ru"}, Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}}}}
	ctx := context.WithValue(context.Background(), contextkey.IsAdminKey, false)
	h.PromoEnterCallbackHandler(ctx, b, upd)
	if !strings.Contains(httpc.body, tm.GetText("ru", "enter_promocode_prompt")) {
		t.Fatalf("unexpected body: %s", httpc.body)
	}
}

func TestUnknownCallbackHandler(t *testing.T) {
	SetTestEnv(t)
	tm := translation.GetInstance()
	_ = tm.InitDefaultTranslations()
	httpc := &stubHTTPPromo{}
	b, _ := bot.New("t", bot.WithHTTPClient(time.Second, httpc), bot.WithSkipGetMe())
	h := handlerpkg.NewHandler(nil, nil, tm, &StubCustomerRepo{}, nil, nil, nil, nil, nil, nil)

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{ID: "1", Data: "unknown_cb", From: models.User{ID: 1, LanguageCode: "ru"}, Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}}}}
	ctx := context.WithValue(context.Background(), contextkey.IsAdminKey, false)
	h.UnknownCallbackHandler(ctx, b, upd)
	if !strings.Contains(httpc.body, tm.GetText("ru", "unknown_callback")) {
		t.Fatalf("unexpected body: %s", httpc.body)
	}
}
