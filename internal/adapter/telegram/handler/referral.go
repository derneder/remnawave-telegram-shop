package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/contextkey"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	"remnawave-tg-shop-bot/internal/service/payment"
	menu "remnawave-tg-shop-bot/internal/ui/menu"
)

func (h *Handler) ReferralCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	langCode := update.CallbackQuery.From.LanguageCode
	tm := translation.GetInstance()
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	_, err = h.findOrCreateCustomer(ctx, update.CallbackQuery.From.ID, langCode)
	if err != nil {
		slog.Error("find or create customer", "err", err)
		return
	}

	var kb [][]models.InlineKeyboardButton
	if contextkey.IsAdminFromContext(ctx) {
		kb = menu.BuildRefPromoAdminMenu(langCode)
	} else {
		kb = menu.BuildRefPromoUserMenu(langCode)
	}

	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msgID,
		ParseMode:   models.ParseModeHTML,
		Text:        tm.GetText(langCode, "referral_menu_text"),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("Error sending referral menu", "err", err)
	}
}

func (h *Handler) PromoEnterCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	langCode := update.CallbackQuery.From.LanguageCode
	tm := translation.GetInstance()

	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	_, err = h.findOrCreateCustomer(ctx, update.CallbackQuery.From.ID, langCode)
	if err != nil {
		slog.Error("find or create customer", "err", err)
		return
	}

	h.expectPromo(update.CallbackQuery.From.ID)

	kb := [][]models.InlineKeyboardButton{
		{
			{Text: tm.GetText(langCode, "back_button"), CallbackData: CallbackReferral},
		},
	}

	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msgID,
		ParseMode:   models.ParseModeHTML,
		Text:        tm.GetText(langCode, "enter_promocode_prompt"),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("Error sending enter_promocode_prompt code msg", "err", err)
	}

}

func (h *Handler) ReferralStatsCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	langCode := update.CallbackQuery.From.LanguageCode
	tm := translation.GetInstance()
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	customer, err := h.findOrCreateCustomer(ctx, update.CallbackQuery.From.ID, langCode)
	if err != nil {
		slog.Error("find or create customer", "err", err)
		return
	}

	// stats are not tracked in current implementation
	invited := 0
	subscribed := 0
	bonusTotal := 0

	refLink := fmt.Sprintf("https://t.me/%s?start=ref_%d", update.CallbackQuery.From.Username, customer.TelegramID)

	text := fmt.Sprintf(tm.GetText(langCode, "referral_system_text"), invited, subscribed, bonusTotal, refLink, config.GetReferralBonus())

	kb := [][]models.InlineKeyboardButton{
		{{Text: tm.GetText(langCode, "invite_friend_button"), URL: refLink}},
		{{Text: tm.GetText(langCode, "back_button"), CallbackData: CallbackReferral}},
	}

	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msgID,
		ParseMode:   models.ParseModeHTML,
		Text:        text,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("Error sending referral stats", "err", err)
	}
}

func (h *Handler) PromocodeCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	lang := update.Message.From.LanguageCode
	tm := translation.GetInstance()
	customer, err := h.findOrCreateCustomer(ctx, update.Message.Chat.ID, lang)
	if err != nil {
		slog.Error("find or create customer", "err", err)
		return
	}

	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: tm.GetText(lang, "promo_invalid")}); err != nil {
			slog.Error("send promo invalid", "err", err)
		}
		return
	}

	code := parts[1]
	promo, err := h.paymentService.ApplyPromocode(ctx, customer, code)
	if err != nil {
		var text string
		if errors.Is(err, payment.ErrPromocodeNotFound) {
			text = tm.GetText(lang, "promo_not_found")
		} else if errors.Is(err, payment.ErrPromocodeExpired) {
			text = tm.GetText(lang, "promo_expired")
		} else if errors.Is(err, payment.ErrPromocodeLimitExced) {
			text = tm.GetText(lang, "promo_limit_reached")
		} else {
			text = tm.GetText(lang, "promo_invalid")
		}
		if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: text}); serr != nil {
			slog.Error("send promo invalid", "err", serr)
		}
		return
	}
	if promo.Type == 2 {
		if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, ParseMode: models.ParseModeHTML, Text: fmt.Sprintf(tm.GetText(lang, "promo_balance_applied"), promo.Amount/100, int(customer.Balance))}); serr != nil {
			slog.Error("send balance promo", "err", serr)
		}
		return
	}

	until := ""
	if customer.ExpireAt != nil {
		until = customer.ExpireAt.Format("02.01.2006 15:04")
	}

	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: fmt.Sprintf(tm.GetText(lang, "promo_applied"), until)}); err != nil {
		slog.Error("send promo applied", "err", err)
	}
}

func (h *Handler) PromoCodeMessageHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !h.consumePromo(update.Message.Chat.ID) {
		return
	}
	lang := update.Message.From.LanguageCode
	tm := translation.GetInstance()

	customer, err := h.findOrCreateCustomer(ctx, update.Message.Chat.ID, lang)
	if err != nil {
		slog.Error("find or create customer", "err", err)
		return
	}

	code := strings.TrimSpace(update.Message.Text)
	promo, err := h.paymentService.ApplyPromocode(ctx, customer, code)
	if err != nil {
		var text string
		if errors.Is(err, payment.ErrPromocodeNotFound) {
			text = tm.GetText(lang, "promo_not_found")
		} else if errors.Is(err, payment.ErrPromocodeExpired) {
			text = tm.GetText(lang, "promo_expired")
		} else if errors.Is(err, payment.ErrPromocodeLimitExced) {
			text = tm.GetText(lang, "promo_limit_reached")
		} else {
			text = tm.GetText(lang, "promo_invalid")
		}
		if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: text}); serr != nil {
			slog.Error("send promo invalid", "err", serr)
		}
		return
	}
	if promo.Type == 2 {
		if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, ParseMode: models.ParseModeHTML, Text: fmt.Sprintf(tm.GetText(lang, "promo_balance_applied"), promo.Amount/100, int(customer.Balance))}); serr != nil {
			slog.Error("send balance promo", "err", serr)
		}
		return
	}

	until := ""
	if customer.ExpireAt != nil {
		until = customer.ExpireAt.Format("02.01.2006 15:04")
	}

	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf(tm.GetText(lang, "promo_applied"), until)}); err != nil {
		slog.Error("send promo applied", "err", err)
	}
}
