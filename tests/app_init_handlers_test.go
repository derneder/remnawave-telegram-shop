package tests

import (
	"context"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnawave-tg-shop-bot/internal/adapter/telegram/handler"
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

func TestInitHandlers(t *testing.T) {
	b, err := bot.New("1:1", bot.WithSkipGetMe(), bot.WithNotAsyncHandlers())
	if err != nil {
		t.Fatal(err)
	}

	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatal(err)
	}

	repo := &StubCustomerRepo{}
       h := handler.NewHandler(nil, nil, tm, repo, nil, nil, nil, nil, nil, nil)

	b.RegisterHandler(bot.HandlerTypeMessageText, "/connect", bot.MatchTypeExact, h.ConnectCommandHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackConnect, bot.MatchTypePrefix, h.ConnectCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)

	upd := &models.Update{
		ID: 1,
		Message: &models.Message{
			Chat:     models.Chat{ID: 1},
			From:     &models.User{ID: 1, LanguageCode: "en"},
			Text:     "/connect",
			Entities: []models.MessageEntity{{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: len("/connect")}},
		},
	}
	b.ProcessUpdate(context.Background(), upd)
	if repo.Calls == 0 {
		t.Fatalf("command handler not executed")
	}

	repo.Calls = 0
	upd = &models.Update{
		ID: 2,
		CallbackQuery: &models.CallbackQuery{
			ID:      "cb",
			From:    models.User{ID: 1, LanguageCode: "en"},
			Data:    handler.CallbackConnect,
			Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}},
		},
	}
	b.ProcessUpdate(context.Background(), upd)
	if repo.Calls == 0 {
		t.Fatalf("callback handler not executed")
	}
}
