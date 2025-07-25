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

type menuClient struct{ body string }

func (c *menuClient) Do(req *http.Request) (*http.Response, error) {
	b, _ := io.ReadAll(req.Body)
	c.body = string(b)
	resp := &http.Response{StatusCode: http.StatusOK}
	resp.Body = io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`))
	return resp, nil
}

func TestPromoCodesCallbackHandler_AdminButtons(t *testing.T) {
	SetTestEnv(t)
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}
	repo := &StubCustomerRepo{}
	client := &menuClient{}
	b, _ := bot.New("t", bot.WithHTTPClient(time.Second, client), bot.WithSkipGetMe())
	h := handlerpkg.NewHandler(nil, nil, tm, repo, nil, nil, nil, nil, nil, nil)
	upd := &models.Update{CallbackQuery: &models.CallbackQuery{ID: "1", From: models.User{ID: 1, LanguageCode: "en"}, Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}}}}
	h.PromoCodesCallbackHandler(context.Background(), b, upd)
	if !strings.Contains(client.body, handlerpkg.CallbackAdminSubPromo) {
		t.Fatal("sub promo button missing")
	}
	if !strings.Contains(client.body, handlerpkg.CallbackAdminBalPromo) {
		t.Fatal("bal promo button missing")
	}
}
