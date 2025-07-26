package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/service/payment"
)

// handlePromoCommand processes promo related commands. If parseArgs is true, it
// will attempt to apply a code passed as argument, otherwise it just prompts the
// user to enter a promo code.
func (h *Handler) handlePromoCommand(ctx context.Context, b *bot.Bot, update *models.Update, parseArgs bool) {
	lang := update.Message.From.LanguageCode
	customer, err := h.findOrCreateCustomer(ctx, update.Message.Chat.ID, lang)
	if err != nil {
		slog.Error("find or create customer", "err", err)
		return
	}

	parts := strings.Fields(update.Message.Text)
	if parseArgs && len(parts) > 1 {
		h.applyPromocode(ctx, b, update.Message.Chat.ID, lang, customer, parts[1])
		return
	}

	h.promptPromoActivation(ctx, b, update.Message.Chat.ID, lang)
}

// promptPromoActivation sets the state so the next text message is treated as a
// promo code and asks user to enter it.
func (h *Handler) promptPromoActivation(ctx context.Context, b *bot.Bot, chatID int64, lang string) {
	h.expectPromo(chatID)
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: h.translation.GetText(lang, "promo.activate.prompt")}); err != nil {
		slog.Error("send promo activate prompt", "err", err)
	}
}

// applyPromocode applies the provided promo code to the customer and replies
// with the result.
func (h *Handler) applyPromocode(ctx context.Context, b *bot.Bot, chatID int64, lang string, customer *domaincustomer.Customer, code string) {
	promo, err := h.paymentService.ApplyPromocode(ctx, customer, code)
	if err != nil {
		var text string
		switch {
		case errors.Is(err, payment.ErrPromocodeNotFound):
			text = h.translation.GetText(lang, "promo_not_found")
		case errors.Is(err, payment.ErrPromocodeExpired):
			text = h.translation.GetText(lang, "promo_expired")
		case errors.Is(err, payment.ErrPromocodeLimitExced):
			text = h.translation.GetText(lang, "promo_limit_reached")
		default:
			text = h.translation.GetText(lang, "promo_invalid")
		}
		if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: text}); serr != nil {
			slog.Error("send promo invalid", "err", serr)
		}
		return
	}
	if promo.Type == 2 {
		if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, ParseMode: models.ParseModeHTML, Text: fmt.Sprintf(h.translation.GetText(lang, "promo_balance_applied"), promo.Amount/100, int(customer.Balance))}); serr != nil {
			slog.Error("send balance promo", "err", serr)
		}
		return
	}

	until := ""
	if customer.ExpireAt != nil {
		until = customer.ExpireAt.Format("02.01.2006 15:04")
	}

	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: fmt.Sprintf(h.translation.GetText(lang, "promo_applied"), until)}); err != nil {
		slog.Error("send promo applied", "err", err)
	}
}
