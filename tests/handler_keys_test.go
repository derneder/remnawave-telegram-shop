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
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

type stubHTTPKeys struct{ body string }

func (h *stubHTTPKeys) Do(r *http.Request) (*http.Response, error) {
	b, _ := io.ReadAll(r.Body)
	h.body = string(b)
	resp := &http.Response{StatusCode: http.StatusOK}
	resp.Body = io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`))
	return resp, nil
}

func TestKeysCallbackHandler_VlessScheme(t *testing.T) {
	SetTestEnv(t)
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}
	httpc := &stubHTTPKeys{}
	b, _ := bot.New("t", bot.WithHTTPClient(time.Second, httpc), bot.WithSkipGetMe())
	link := "vless://example.com:443?encryption=none#test"
	repo := &StubCustomerRepo{CustomerByTelegramID: &domaincustomer.Customer{TelegramID: 1, SubscriptionLink: &link}}
	h := handlerpkg.NewHandler(nil, nil, tm, repo, nil, nil, nil, nil, nil, nil)

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{From: models.User{ID: 1, LanguageCode: "en"}, Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}}}}

	h.KeysCallbackHandler(context.Background(), b, upd)

	if !strings.Contains(httpc.body, link) {
		t.Fatalf("expected link in body, got %s", httpc.body)
	}
}
