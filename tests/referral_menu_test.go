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

type stubHTTP2 struct{ body string }

func (h *stubHTTP2) Do(r *http.Request) (*http.Response, error) {
	b, _ := io.ReadAll(r.Body)
	h.body = string(b)
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
	if !strings.Contains(httpc.body, string(uimenu.CallbackPromoUserActivate)) {
		t.Fatalf("user menu missing activate button")
	}

	httpc.body = ""
	ctx = context.WithValue(context.Background(), contextkey.IsAdminKey, true)
	h.ReferralCallbackHandler(ctx, b, upd)
	if !strings.Contains(httpc.body, string(uimenu.CallbackPromoAdminMenu)) {
		t.Fatalf("admin menu missing admin panel")
	}
}
