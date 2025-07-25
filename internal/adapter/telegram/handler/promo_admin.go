package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

func (h *Handler) AddSubPromoCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !config.IsAdmin(update.Message.Chat.ID) {
		return
	}
	parts := strings.Fields(update.Message.Text)
	if len(parts) != 4 {
		return
	}
	days, err1 := strconv.Atoi(parts[2])
	limit, err2 := strconv.Atoi(parts[3])
	if err1 != nil || err2 != nil || days <= 0 || limit < 0 {
		return
	}
	if err := h.promotionService.CreateSubscription(ctx, parts[1], days, limit, update.Message.Chat.ID); err != nil {
		slog.Error("create sub promo", "err", err)
	}
}

func (h *Handler) AddBalPromoCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !config.IsAdmin(update.Message.Chat.ID) {
		return
	}
	parts := strings.Fields(update.Message.Text)
	if len(parts) != 3 {
		return
	}
	amountRub, err1 := strconv.Atoi(parts[1])
	limit, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || amountRub <= 0 || limit < 0 {
		return
	}
	code, err := h.promotionService.CreateBalance(ctx, amountRub*100, limit, update.Message.Chat.ID)
	if err != nil {
		slog.Error("create bal promo", "err", err)
		return
	}
	kb := [][]models.InlineKeyboardButton{{{Text: "\xF0\x9F\x92\x8E PROMO-\xD0\x9A\xD0\x9E\xD0\x94", CallbackData: "PROMO_" + code}}}
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: fmt.Sprintf("promo %s", code), ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}})
}

func (h *Handler) AdminSubPromoCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !config.IsAdmin(update.CallbackQuery.From.ID) {
		return
	}
	h.expectSubPromo(update.CallbackQuery.From.ID)
	tm := translation.GetInstance()
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID})
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.CallbackQuery.From.ID, Text: tm.GetText(update.CallbackQuery.From.LanguageCode, "admin_subpromo_prompt")})
}

func (h *Handler) AdminBalPromoCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !config.IsAdmin(update.CallbackQuery.From.ID) {
		return
	}
	h.expectBalPromo(update.CallbackQuery.From.ID)
	tm := translation.GetInstance()
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID})
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.CallbackQuery.From.ID, Text: tm.GetText(update.CallbackQuery.From.LanguageCode, "admin_balpromo_prompt")})
}

func (h *Handler) AdminSubPromoMessageHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !h.consumeSubPromo(update.Message.Chat.ID) {
		return
	}
	if !config.IsAdmin(update.Message.Chat.ID) {
		return
	}
	parts := strings.Fields(update.Message.Text)
	if len(parts) != 3 {
		return
	}
	days, err1 := strconv.Atoi(parts[1])
	limit, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || days <= 0 || limit < 0 {
		return
	}
	if err := h.promotionService.CreateSubscription(ctx, parts[0], days, limit, update.Message.Chat.ID); err != nil {
		slog.Error("create admin sub promo", "err", err)
		return
	}
	tm := translation.GetInstance()
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: fmt.Sprintf(tm.GetText(update.Message.From.LanguageCode, "admin_subpromo_done"), parts[0])})
}

func (h *Handler) AdminBalPromoMessageHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !h.consumeBalPromo(update.Message.Chat.ID) {
		return
	}
	if !config.IsAdmin(update.Message.Chat.ID) {
		return
	}
	parts := strings.Fields(update.Message.Text)
	if len(parts) != 2 {
		return
	}
	amountRub, err1 := strconv.Atoi(parts[0])
	limit, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || amountRub <= 0 || limit < 0 {
		return
	}
	code, err := h.promotionService.CreateBalance(ctx, amountRub*100, limit, update.Message.Chat.ID)
	if err != nil {
		slog.Error("create admin bal promo", "err", err)
		return
	}
	tm := translation.GetInstance()
	kb := [][]models.InlineKeyboardButton{{{Text: "\xF0\x9F\x92\x8E PROMO-\xD0\x9A\xD0\x9E\xD0\x94", CallbackData: "PROMO_" + code}}}
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: fmt.Sprintf(tm.GetText(update.Message.From.LanguageCode, "admin_balpromo_done"), code), ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}})
}
