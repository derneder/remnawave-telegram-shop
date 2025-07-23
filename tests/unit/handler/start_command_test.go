package handler_test

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
	"remnawave-tg-shop-bot/tests/testutils"
)

type startHTTPClient struct{}

func (c *startHTTPClient) Do(req *http.Request) (*http.Response, error) {
	resp := &http.Response{StatusCode: http.StatusOK}
	resp.Body = io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`))
	return resp, nil
}

func TestStartCommandHandler_NoArgs(t *testing.T) {
	trans := translation.GetInstance()
	if err := trans.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}

	repo := &testutils.StubCustomerRepo{}
	h := handlerpkg.NewHandler(nil, nil, trans, repo, nil, nil, nil, nil, nil)

	b, err := bot.New("token", bot.WithHTTPClient(time.Second, &startHTTPClient{}), bot.WithSkipGetMe())
	if err != nil {
		t.Fatalf("bot init: %v", err)
	}

	upd := &models.Update{
		Message: &models.Message{
			Chat:     models.Chat{ID: 1},
			From:     &models.User{ID: 1, LanguageCode: "en", FirstName: "u"},
			Text:     "/start",
			Entities: []models.MessageEntity{{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: len("/start")}},
		},
	}

	ctx := context.WithValue(context.Background(), ctxKey{}, "v")
	h.StartCommandHandler(ctx, b, upd)

	if repo.Ctx.Value(ctxKey{}) != "v" {
		t.Errorf("context not propagated")
	}
}
