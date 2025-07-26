package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnawave-tg-shop-bot/internal/pkg/contextkey"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	pg "remnawave-tg-shop-bot/internal/repository/pg"
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

	admin := contextkey.IsAdminFromContext(ctx)
	kb := menu.BuildPromoRefMain(langCode, admin)

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
		Text:        tm.GetText(langCode, "promo.activate.prompt"),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("Error sending promo activate prompt", "err", err)
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

	text, kb := h.buildReferralInfo(customer, langCode)
	kb = append(kb, []models.InlineKeyboardButton{{Text: tm.GetText(langCode, "back_button"), CallbackData: CallbackReferral}})

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

func (h *Handler) PromoMyListCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	lang := update.CallbackQuery.From.LanguageCode
	tm := translation.GetInstance()
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	promos, err := h.promocodeRepository.FindByCreator(ctx, update.CallbackQuery.From.ID)
	if err != nil {
		slog.Error("list promos", "err", err)
		return
	}

	var filtered []pg.Promocode
	for _, p := range promos {
		if p.Deleted {
			continue
		}
		if p.UsesLeft == 0 {
			continue
		}
		filtered = append(filtered, p)
	}
	promos = filtered

	var text strings.Builder
	var kb [][]models.InlineKeyboardButton

	if len(promos) == 0 {
		text.WriteString(tm.GetText(lang, "promo.list.empty"))
	} else {
		text.WriteString(tm.GetText(lang, "promo.list.title"))
		text.WriteString("\n\n")
		for _, p := range promos {
			text.WriteString(buildPromoItemText(lang, p))
			text.WriteString("\n")
			kb = append(kb, buildPromoItemButtons(lang, p))
		}
	}

	kb = append(kb, []models.InlineKeyboardButton{{Text: tm.GetText(lang, "back_button"), CallbackData: CallbackReferral}})

	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msgID,
		ParseMode:   models.ParseModeHTML,
		Text:        text.String(),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("send promo list", "err", err)
	}
}

func (h *Handler) PromoMyFreezeCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	idStr := strings.TrimPrefix(update.CallbackQuery.Data, menu.CallbackPromoMyFreeze+":")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.promotionService.Freeze(ctx, id); err != nil {
		slog.Error("freeze promo", "err", err)
	}
	h.PromoMyListCallbackHandler(ctx, b, update)
}

func (h *Handler) PromoMyUnfreezeCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	idStr := strings.TrimPrefix(update.CallbackQuery.Data, menu.CallbackPromoMyUnfreeze+":")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.promotionService.Unfreeze(ctx, id); err != nil {
		slog.Error("unfreeze promo", "err", err)
	}
	h.PromoMyListCallbackHandler(ctx, b, update)
}

func (h *Handler) PromoMyDeleteCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	lang := update.CallbackQuery.From.LanguageCode
	tm := translation.GetInstance()
	idStr := strings.TrimPrefix(update.CallbackQuery.Data, menu.CallbackPromoMyDelete+":")
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	kb := [][]models.InlineKeyboardButton{
		{{Text: tm.GetText(lang, "confirm_button"), CallbackData: menu.CallbackPromoMyDeleteConfirm + ":" + idStr}},
		{{Text: tm.GetText(lang, "cancel_button"), CallbackData: menu.CallbackPromoMyList}},
	}
	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msgID,
		ParseMode:   models.ParseModeHTML,
		Text:        tm.GetText(lang, "promo.delete.confirm"),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("send delete confirm", "err", err)
	}
}

func (h *Handler) PromoMyDeleteConfirmCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	idStr := strings.TrimPrefix(update.CallbackQuery.Data, menu.CallbackPromoMyDeleteConfirm+":")
	lang := update.CallbackQuery.From.LanguageCode
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.promotionService.Delete(ctx, id); err != nil {
		slog.Error("delete promo", "err", err)
	}
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID, Text: translation.GetInstance().GetText(lang, "promo.delete.success")})
	h.PromoMyListCallbackHandler(ctx, b, update)
}

func (h *Handler) PromocodeCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.handlePromoCommand(ctx, b, update, false)
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

func buildPromoItemText(lang string, p pg.Promocode) string {
	tm := translation.GetInstance()
	icon := tm.GetText(lang, "promo.item.status_icon.active")
	if !p.Active {
		icon = tm.GetText(lang, "promo.item.status_icon.inactive")
	}

	if p.Type == 2 {
		return fmt.Sprintf(
			tm.GetText(lang, "promo.item.compact_bal"),
			icon,
			p.Code,
			p.Amount/100,
			p.UsesLeft,
		)
	}

	months := p.Months
	if months == 0 && p.Days > 0 {
		months = p.Days / 30
	}
	if months == 0 {
		months = 1
	}

	return fmt.Sprintf(
		tm.GetText(lang, "promo.item.compact_sub"),
		icon,
		p.Code,
		months,
		p.UsesLeft,
	)
}

func buildPromoItemButtons(lang string, p pg.Promocode) []models.InlineKeyboardButton {
	tm := translation.GetInstance()
	idStr := strconv.FormatInt(p.ID, 10)
	var row []models.InlineKeyboardButton
	if p.Active {
		row = append(row, models.InlineKeyboardButton{Text: tm.GetText(lang, "freeze_button"), CallbackData: menu.CallbackPromoMyFreeze + ":" + idStr})
	} else {
		row = append(row, models.InlineKeyboardButton{Text: tm.GetText(lang, "unfreeze_button"), CallbackData: menu.CallbackPromoMyUnfreeze + ":" + idStr})
	}
	row = append(row, models.InlineKeyboardButton{Text: tm.GetText(lang, "delete_button"), CallbackData: menu.CallbackPromoMyDelete + ":" + idStr})
	return row
}
