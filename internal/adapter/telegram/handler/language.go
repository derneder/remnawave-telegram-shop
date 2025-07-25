package handler

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// LanguageCallbackHandler shows language selection buttons.
func (h *Handler) LanguageCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	lang := update.CallbackQuery.From.LanguageCode
	kb := [][]models.InlineKeyboardButton{
		{{Text: h.translation.GetText(lang, "language_ru"), CallbackData: CallbackSetLanguage + "?lang=ru"}},
		{{Text: h.translation.GetText(lang, "language_en"), CallbackData: CallbackSetLanguage + "?lang=en"}},
		{{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackOther}},
	}
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msgID,
		ParseMode:   models.ParseModeHTML,
		Text:        h.translation.GetText(lang, "choose_language_text"),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("send language menu", "err", err)
	}
}

// SetLanguageCallbackHandler updates user language.
func (h *Handler) SetLanguageCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	data := parseCallbackData(update.CallbackQuery.Data)
	lang := data["lang"]
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	customer, err := h.findOrCreateCustomer(ctx, update.CallbackQuery.From.ID, lang)
	if err != nil {
		slog.Error("find or create customer", "err", err)
		return
	}
	if err := h.customerRepository.UpdateFields(ctx, customer.ID, map[string]interface{}{"language": lang}); err != nil {
		slog.Error("update language", "err", err)
		return
	}
	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msgID,
		ParseMode:   models.ParseModeHTML,
		Text:        h.translation.GetText(lang, "language_changed"),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{{{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackOther}}}},
	})
	if err != nil {
		slog.Error("send language updated", "err", err)
	}
}
