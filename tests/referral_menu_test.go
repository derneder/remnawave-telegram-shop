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
)

type stubHTTP2 struct{ bodies []string }

func (h *stubHTTP2) Do(r *http.Request) (*http.Response, error) {
	b, _ := io.ReadAll(r.Body)
	h.bodies = append(h.bodies, string(b))
	resp := &http.Response{StatusCode: http.StatusOK}
	resp.Body = io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`))
	return resp, nil
}

func TestReferralCallbackHandler_UserAdmin(t *testing.T) {
	SetTestEnv(t)
	tm := translation.GetInstance()
	_ = tm.InitDefaultTranslations()
	httpc := &stubHTTP2{}
	b, _ := bot.New("t", bot.WithHTTPClient(time.Second, httpc), bot.WithSkipGetMe())
	h := handlerpkg.NewHandler(nil, nil, tm, &StubCustomerRepo{}, nil, nil, nil, nil, nil, nil)

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{ID: "1", From: models.User{ID: 1, LanguageCode: "ru"}, Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}}}}

	ctx := context.WithValue(context.Background(), contextkey.IsAdminKey, false)
	h.ReferralCallbackHandler(ctx, b, upd)
	if len(httpc.bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(httpc.bodies))
	}
	if !strings.Contains(httpc.bodies[0], tm.GetText("ru", "promo_ref_menu_text")) {
		t.Fatalf("menu not sent")
	}
	if !strings.Contains(httpc.bodies[1], "callback_query_id") {
		t.Fatalf("callback not answered")
	}
}
