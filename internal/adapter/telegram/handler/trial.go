package handler

import (
	"context"

	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/contextkey"
	"remnawave-tg-shop-bot/internal/ui"
	"remnawave-tg-shop-bot/utils"
)

func (h *Handler) TrialCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if config.TrialDays() == 0 {
		return
	}
	c, err := h.customerRepository.FindByTelegramId(ctx, update.CallbackQuery.From.ID)
	if err != nil {
		slog.Error("Error finding customer", "err", err)
		return
	}
	if c == nil {
		slog.Error("customer not exist", "telegramId", utils.MaskHalfInt64(update.CallbackQuery.From.ID), "error", err)
		return
	}
	if c.SubscriptionLink != nil {
		return
	}
	callback := update.CallbackQuery.Message.Message
	langCode := update.CallbackQuery.From.LanguageCode
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    callback.Chat.ID,
		MessageID: callback.ID,
		Text:      h.translation.GetText(langCode, "trial_text"),
		ParseMode: models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: h.translation.GetText(langCode, "activate_trial_button"), CallbackData: CallbackActivateTrial}},
			{{Text: h.translation.GetText(langCode, "back_to_account_button"), CallbackData: CallbackStart}},
		}},
	})
	if err != nil {
		slog.Error("Error sending /trial message", "err", err)
	}
}

func (h *Handler) ActivateTrialCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if config.TrialDays() == 0 {
		return
	}
	c, err := h.customerRepository.FindByTelegramId(ctx, update.CallbackQuery.From.ID)
	if err != nil {
		slog.Error("Error finding customer", "err", err)
		return
	}
	if c == nil {
		slog.Error("customer not exist", "telegramId", utils.MaskHalfInt64(update.CallbackQuery.From.ID), "error", err)
		return
	}
	if c.SubscriptionLink != nil {
		return
	}
	callback := update.CallbackQuery.Message.Message
	ctxWithUsername := context.WithValue(ctx, contextkey.Username, update.CallbackQuery.From.Username)
	_, err = h.paymentService.ActivateTrial(ctxWithUsername, update.CallbackQuery.From.ID)
	if err != nil {
		slog.Error("Error activate trial", "err", err)
	}

	langCode := update.CallbackQuery.From.LanguageCode
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      callback.Chat.ID,
		MessageID:   callback.ID,
		Text:        h.translation.GetText(langCode, "trial_activated"),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: ui.ConnectKeyboard(langCode, "back_to_account_button", CallbackStart)},
	})
	if err != nil {
		slog.Error("Error sending /trial message", "err", err)
	}
}
