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

	"remnawave-tg-shop-bot/internal/pkg/config"
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

	// stats are not tracked in current implementation
	invited := 0
	subscribed := 0
	bonusTotal := 0

	botURL := strings.TrimPrefix(config.BotURL(), "https://t.me/")
	botURL = strings.TrimPrefix(botURL, "http://t.me/")
	refLink := menu.BuildReferralLink(botURL, fmt.Sprintf("ref_%d", customer.TelegramID))

	var sb strings.Builder
	sb.WriteString(tm.GetText(langCode, "ref.msg.welcome"))
	sb.WriteString("\n\n")
	sb.WriteString(tm.GetText(langCode, "ref.msg.stats_title"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(tm.GetText(langCode, "ref.msg.stats_invited"), invited))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(tm.GetText(langCode, "ref.msg.stats_paid"), subscribed))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(tm.GetText(langCode, "ref.msg.stats_sum"), bonusTotal))
	sb.WriteString("\n\n")
	sb.WriteString(tm.GetText(langCode, "ref.msg.link_title"))
	sb.WriteString("\n")
	sb.WriteString(tm.GetText(langCode, "ref.msg.link_note"))
	sb.WriteString("\n")
	sb.WriteString(tm.GetText(langCode, "ref.msg.copy_hint"))
	sb.WriteString(" ")
	sb.WriteString(refLink)
	sb.WriteString("\n\n")
	sb.WriteString(tm.GetText(langCode, "ref.msg.bonus_info_title"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(tm.GetText(langCode, "ref.msg.bonus_info_text"), config.GetReferralBonus(), config.GetReferralBonus()))

	text := sb.String()

	kb := [][]models.InlineKeyboardButton{
		{{Text: tm.GetText(langCode, "ref.button.invite"), URL: refLink}},
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

	var text strings.Builder
	var kb [][]models.InlineKeyboardButton

	if len(promos) == 0 {
		text.WriteString(tm.GetText(lang, "promo.list.empty"))
	} else {
		text.WriteString(tm.GetText(lang, "promo.list.title"))
		text.WriteString("\n\n")
		for _, p := range promos {
			text.WriteString(buildPromoItemText(lang, p))
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
       id, _ := strconv.ParseInt(idStr, 10, 64)
       if err := h.promotionService.Delete(ctx, id); err != nil {
               slog.Error("delete promo", "err", err)
       }
	h.PromoMyListCallbackHandler(ctx, b, update)
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
		h.expectPromo(update.Message.Chat.ID)
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: tm.GetText(lang, "promo.activate.prompt")}); err != nil {
			slog.Error("send promo activate prompt", "err", err)
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

func buildPromoItemText(lang string, p pg.Promocode) string {
	tm := translation.GetInstance()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(tm.GetText(lang, "promo.item.code"), p.Code))
	sb.WriteString("\n")
	t := tm.GetText(lang, "promo.item.type.subscription")
	term := ""
	if p.Type == 2 {
		t = tm.GetText(lang, "promo.item.type.balance")
		term = fmt.Sprintf(tm.GetText(lang, "promo.item.term.amount"), p.Amount/100)
	} else {
		days := p.Days
		if days == 0 {
			days = p.Months * 30
		}
		term = fmt.Sprintf(tm.GetText(lang, "promo.item.term.days"), days)
	}
	sb.WriteString(fmt.Sprintf(tm.GetText(lang, "promo.item.type"), t))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(tm.GetText(lang, "promo.item.term"), term))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(tm.GetText(lang, "promo.item.uses"), p.UsesLeft))
	sb.WriteString("\n")
	status := tm.GetText(lang, "promo.item.status.active")
	if !p.Active {
		status = tm.GetText(lang, "promo.item.status.inactive")
	}
	sb.WriteString(fmt.Sprintf(tm.GetText(lang, "promo.item.status"), status))
	sb.WriteString("\n\n")
	return sb.String()
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
