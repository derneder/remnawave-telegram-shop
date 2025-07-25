package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/translation"
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

	kb := [][]models.InlineKeyboardButton{
		{
			{Text: tm.GetText(langCode, "enter_promocode_button"), CallbackData: CallbackPromoEnter},
		},
		{
			{Text: tm.GetText(langCode, "referral_system_button"), CallbackData: CallbackReferralStats},
		},
		{
			{Text: tm.GetText(langCode, "personal_codes_button"), CallbackData: CallbackPromoCodes},
		},
		{
			{Text: tm.GetText(langCode, "back_to_account_button"), CallbackData: CallbackStart},
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
		Text:        tm.GetText(langCode, "referral_menu_text"),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("Error sending referral menu", "err", err)
	}
}

func (h *Handler) PromoCreateCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
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
	data := parseCallbackData(update.CallbackQuery.Data)
	monthStr := data["m"]
	usesStr := data["u"]
	msg := update.CallbackQuery.Message.Message
	if msg == nil {
		slog.Error("callback message missing")
		return
	}
	if usesStr == "" {
		h.promptPromoUses(ctx, b, msg, langCode)
		return
	}
	uses, _ := strconv.Atoi(usesStr)
	if monthStr == "" {
		h.promptPromoMonths(ctx, b, msg, langCode, uses, customer)
		return
	}

	month, _ := strconv.Atoi(monthStr)
	code, err := h.paymentService.CreatePromocode(ctx, customer, month, uses)

	kb := [][]models.InlineKeyboardButton{
		{
			{Text: tm.GetText(langCode, "back_button"), CallbackData: CallbackPromoCodes},
		},
	}

	if err != nil {
		var curMsg *models.Message
		if update.CallbackQuery.Message.Message != nil {
			curMsg = update.CallbackQuery.Message.Message
		}
		_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   msgID,
			ParseMode:   models.ParseModeHTML,
			Text:        tm.GetText(langCode, "insufficient_balance"),
			ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
		})
		if err != nil {
			slog.Error("Error sending insufficient_balance code msg", "err", err)
		}

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
		Text:        fmt.Sprintf(tm.GetText(langCode, "promocode_created"), code),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("Error sending succesfully_created code msg", "err", err)
	}

	slog.Info("promocode created", "code", code, "customer", customer.TelegramID)
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

func (h *Handler) PromoCodesCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
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

	kb := [][]models.InlineKeyboardButton{
		{{Text: tm.GetText(langCode, "create_promocode_button"), CallbackData: CallbackPromoCreate}},
	}
	if config.IsAdmin(update.CallbackQuery.From.ID) {
		kb = append(kb,
			[]models.InlineKeyboardButton{{Text: tm.GetText(langCode, "admin_subpromo_button"), CallbackData: CallbackAdminSubPromo}},
			[]models.InlineKeyboardButton{{Text: tm.GetText(langCode, "admin_balpromo_button"), CallbackData: CallbackAdminBalPromo}},
		)
	}
	kb = append(kb,
		[]models.InlineKeyboardButton{{Text: tm.GetText(langCode, "promo_list_button"), CallbackData: CallbackPromoList}},
		[]models.InlineKeyboardButton{{Text: tm.GetText(langCode, "back_button"), CallbackData: CallbackReferral}},
	)

	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msgID,
		ParseMode:   models.ParseModeHTML,
		Text:        tm.GetText(langCode, "personal_codes_text"),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("Error sending promo codes menu", "err", err)
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
	if err := h.paymentService.ApplyPromocode(ctx, customer, code); err != nil {
		if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: tm.GetText(lang, "promo_invalid")}); serr != nil {
			slog.Error("send promo invalid", "err", serr)
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

func (h *Handler) PromoListCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
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

	codes, err := h.promocodeRepository.FindByCreator(ctx, customer.TelegramID)
	if err != nil {
		slog.Error("error getting promocodes", "err", err)
		return
	}

	var textBuilder strings.Builder
	textBuilder.WriteString(tm.GetText(langCode, "promo_codes_list_intro"))
	if len(codes) == 0 {
		textBuilder.WriteString("\n\n-")
	}
	var kb [][]models.InlineKeyboardButton
	for _, c := range codes {
		used, _ := h.promocodeUsageRepository.CountByPromocodeID(ctx, c.ID)
		total := used + c.UsesLeft
		status := tm.GetText(langCode, "promo_status_active")
		if !c.Active {
			status = tm.GetText(langCode, "promo_status_frozen")
		}
		if c.Type == 2 {
			textBuilder.WriteString(fmt.Sprintf("\n%s — %d ₽ — осталось %d/%d — %s", c.Code, c.Amount/100, c.UsesLeft, total, status))
		} else {
			days := c.Days
			if days == 0 {
				days = c.Months * 30
			}
			textBuilder.WriteString(fmt.Sprintf("\n%s — %d d — осталось %d/%d — %s", c.Code, days, c.UsesLeft, total, status))
		}
		if c.Active {
			kb = append(kb, []models.InlineKeyboardButton{
				{Text: tm.GetText(langCode, "promo_freeze_button"), CallbackData: fmt.Sprintf("%s:%d", CallbackPromoFreeze, c.ID)},
				{Text: tm.GetText(langCode, "promo_delete_button"), CallbackData: fmt.Sprintf("%s:%d", CallbackPromoConfirmationDelete, c.ID)},
			})
		} else {
			kb = append(kb, []models.InlineKeyboardButton{
				{Text: tm.GetText(langCode, "promo_unfreeze_button"), CallbackData: fmt.Sprintf("%s:%d", CallbackPromoUnfreeze, c.ID)},
				{Text: tm.GetText(langCode, "promo_delete_button"), CallbackData: fmt.Sprintf("%s:%d", CallbackPromoConfirmationDelete, c.ID)},
			})
		}
	}

	kb = append(kb, []models.InlineKeyboardButton{{Text: tm.GetText(langCode, "back_button"), CallbackData: CallbackPromoCodes}})

	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msgID,
		ParseMode:   models.ParseModeHTML,
		Text:        textBuilder.String(),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("Error sending promocode list", "err", err)
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
	if err := h.paymentService.ApplyPromocode(ctx, customer, code); err != nil {
		if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: tm.GetText(lang, "promo_invalid")}); serr != nil {
			slog.Error("send promo invalid", "err", serr)
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

func (h *Handler) PromoFreezeCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	parts := strings.Split(update.CallbackQuery.Data, ":")
	if len(parts) != 2 {
		return
	}
	id, _ := strconv.ParseInt(parts[1], 10, 64)
	if err := h.paymentService.SetPromocodeStatus(ctx, id, false); err != nil {
		slog.Error("freeze promocode", "err", err)
	}
	h.PromoListCallbackHandler(ctx, b, update)
}

func (h *Handler) PromoUnfreezeCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	parts := strings.Split(update.CallbackQuery.Data, ":")
	if len(parts) != 2 {
		return
	}
	id, _ := strconv.ParseInt(parts[1], 10, 64)
	if err := h.paymentService.SetPromocodeStatus(ctx, id, true); err != nil {
		slog.Error("unfreeze promocode", "err", err)
	}
	h.PromoListCallbackHandler(ctx, b, update)
}

func (h *Handler) PromoDeleteConfirmationCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	langCode := update.CallbackQuery.From.LanguageCode
	tm := translation.GetInstance()

	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	parts := strings.Split(update.CallbackQuery.Data, ":")
	if len(parts) != 2 {
		return
	}
	id, _ := strconv.ParseInt(parts[1], 10, 64)

	promo, err := h.promocodeRepository.GetById(ctx, id)
	if err != nil {
		slog.Error("failed to get promo", "err", err)

		return
	}

	if promo.Active {
		kb := [][]models.InlineKeyboardButton{
			{{Text: tm.GetText(langCode, "back_button"), CallbackData: CallbackPromoList}},
		}
		var curMsg *models.Message
		if update.CallbackQuery.Message.Message != nil {
			curMsg = update.CallbackQuery.Message.Message
		}
		_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   msgID,
			ParseMode:   models.ParseModeHTML,
			Text:        tm.GetText(langCode, "promo_active_when_delete"),
			ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
		})
		if err != nil {
			slog.Error("Error sending promo_active_when_delete code msg", "err", err)
		}
		return
	}

	kb := [][]models.InlineKeyboardButton{
		{{Text: tm.GetText(langCode, "promo_delete_button"), CallbackData: fmt.Sprintf("%s:%d", CallbackPromoDelete, promo.ID)}},
		{{Text: tm.GetText(langCode, "back_button"), CallbackData: CallbackPromoList}},
	}

	var curMsg2 *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg2 = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg2, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msgID,
		ParseMode:   models.ParseModeHTML,
		Text:        tm.GetText(langCode, "promo_confirm_when_delete"),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("Error sending promo_confirm_when_delete code msg", "err", err)
	}

}

func (h *Handler) PromoDeleteCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	parts := strings.Split(update.CallbackQuery.Data, ":")
	if len(parts) != 2 {
		return
	}
	id, _ := strconv.ParseInt(parts[1], 10, 64)
	if err := h.promocodeRepository.UpdateDeleteStatus(ctx, id, true); err != nil {
		slog.Error("delete promocode", "err", err)
	}
	// Optionally mark deleted: we set active=false
	h.PromoListCallbackHandler(ctx, b, update)
}
