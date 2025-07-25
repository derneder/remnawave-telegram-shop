package handler

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnawave-tg-shop-bot/internal/pkg/translation"
)

func (h *Handler) UnknownCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	data := update.CallbackQuery.Data
	slog.Warn("unknown callback", "data", data)
	lang := update.CallbackQuery.From.LanguageCode
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	tm := translation.GetInstance()
	_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: msgID,
		Text:      tm.GetText(lang, "unknown_callback"),
	})
}
