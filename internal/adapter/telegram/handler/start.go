package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/utils"
)

func (h *Handler) StartCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	ctxWithTime, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	langCode := update.Message.From.LanguageCode
	existingCustomer, err := h.customerRepository.FindByTelegramId(ctx, update.Message.Chat.ID)
	if err != nil {
		slog.Error("error finding customer by telegram id", "err", err)
		return
	}

	if existingCustomer == nil {
		existingCustomer, err = h.customerRepository.Create(ctxWithTime, &domaincustomer.Customer{
			TelegramID: update.Message.Chat.ID,
			Language:   langCode,
			Balance:    0,
		})
		if err != nil {
			slog.Error("error creating customer", "err", err)
			return
		}

		parts := strings.Fields(update.Message.Text)
		if len(parts) > 1 && strings.HasPrefix(parts[1], "ref_") {
			code := strings.TrimPrefix(parts[1], "ref_")
			referrerId, err := strconv.ParseInt(code, 10, 64)
			if err != nil {
				slog.Error("error parsing referrer id", "err", err)
				return
			}
			_, err = h.customerRepository.FindByTelegramId(ctx, referrerId)
			if err == nil {
				_, err := h.referralRepository.Create(ctx, referrerId, existingCustomer.TelegramID)
				if err != nil {
					slog.Error("error creating referral", "err", err)
					return
				}
				slog.Info("referral created", "referrerId", utils.MaskHalfInt64(referrerId), "refereeId", utils.MaskHalfInt64(existingCustomer.TelegramID))
			}
		}
	} else {
		updates := map[string]interface{}{
			"language": langCode,
		}

		err = h.customerRepository.UpdateFields(ctx, existingCustomer.ID, updates)
		if err != nil {
			slog.Error("Error updating customer", "err", err)
			return
		}
	}

	startKb := [][]models.InlineKeyboardButton{{{Text: h.translation.GetText(langCode, "account_button"), CallbackData: CallbackStart}}}
	if config.ChannelURL() != "" {
		startKb = append(startKb, []models.InlineKeyboardButton{{Text: h.translation.GetText(langCode, "channel_button"), URL: config.ChannelURL()}})
	}

	m, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "🧹",
		ReplyMarkup: models.ReplyKeyboardRemove{RemoveKeyboard: true},
	})

	if err != nil {
		slog.Error("Error sending removing reply keyboard", "err", err)
		return
	}

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    update.Message.Chat.ID,
		MessageID: m.ID,
	})

	if err != nil {
		slog.Error("Error deleting message", "err", err)
		return
	}

	text := fmt.Sprintf(h.translation.GetText(langCode, "start_menu_text"), update.Message.From.FirstName)
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: startKb},
		Text:        text,
	})
	if err != nil {
		slog.Error("Error sending /start message", "err", err)
	}
}

func (h *Handler) StartCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	ctxWithTime, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	callback := update.CallbackQuery
	langCode := callback.From.LanguageCode

	existingCustomer, err := h.customerRepository.FindByTelegramId(ctxWithTime, callback.From.ID)
	if err != nil {
		slog.Error("error finding customer by telegram id", "err", err)
		return
	}

	inlineKeyboard := h.buildStartKeyboard(existingCustomer, langCode)

	text := fmt.Sprintf(h.translation.GetText(langCode, "account_menu_text"), callback.From.FirstName) + "\n\n" + h.buildAccountInfo(ctxWithTime, existingCustomer, langCode)
	params := &bot.EditMessageTextParams{
		ChatID:      callback.Message.Message.Chat.ID,
		MessageID:   callback.Message.Message.ID,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboard},
		Text:        text,
	}
	var curMsg *models.Message
	if callback.Message.Message != nil {
		curMsg = callback.Message.Message
	}
	_, err = SafeEditMessageText(ctxWithTime, b, curMsg, params)
	if err != nil {
		slog.Error("Error sending /start message", "err", err)
	}
}

func (h *Handler) resolveConnectButton(lang string) []models.InlineKeyboardButton {
	var inlineKeyboard []models.InlineKeyboardButton

	if config.GetMiniAppURL() != "" {
		inlineKeyboard = []models.InlineKeyboardButton{
			{Text: h.translation.GetText(lang, "connect_button"), WebApp: &models.WebAppInfo{
				URL: config.GetMiniAppURL(),
			}},
		}
	} else {
		inlineKeyboard = []models.InlineKeyboardButton{
			{Text: h.translation.GetText(lang, "connect_button"), CallbackData: CallbackConnect},
		}
	}
	return inlineKeyboard
}

func (h *Handler) buildStartKeyboard(existingCustomer *domaincustomer.Customer, langCode string) [][]models.InlineKeyboardButton {
	var kb [][]models.InlineKeyboardButton

	if existingCustomer.SubscriptionLink == nil && config.TrialDays() > 0 {
		kb = append(kb, []models.InlineKeyboardButton{{Text: h.translation.GetText(langCode, "trial_button"), CallbackData: CallbackTrial}})
	}

	kb = append(kb, []models.InlineKeyboardButton{{Text: h.translation.GetText(langCode, "refresh_button"), CallbackData: CallbackStart}})
	kb = append(kb, []models.InlineKeyboardButton{{Text: h.translation.GetText(langCode, "balance_menu_button"), CallbackData: CallbackBalance}})

	if existingCustomer.SubscriptionLink != nil && existingCustomer.ExpireAt.After(time.Now()) {
		kb = append(kb, h.resolveConnectButton(langCode))
	}

	kb = append(kb, []models.InlineKeyboardButton{{Text: h.translation.GetText(langCode, "referral_button"), CallbackData: CallbackReferral}})
	kb = append(kb, []models.InlineKeyboardButton{{Text: h.translation.GetText(langCode, "other_button"), CallbackData: CallbackOther}})

	var row []models.InlineKeyboardButton
	if config.SupportURL() != "" {
		row = append(row, models.InlineKeyboardButton{Text: h.translation.GetText(langCode, "support_button"), URL: config.SupportURL()})
	}
	if config.ChannelURL() != "" {
		row = append(row, models.InlineKeyboardButton{Text: h.translation.GetText(langCode, "channel_button"), URL: config.ChannelURL()})
	}
	if len(row) > 0 {
		kb = append(kb, row)
	}

	return kb
}

func (h *Handler) buildAccountInfo(ctx context.Context, customer *domaincustomer.Customer, lang string) string {
	user, _ := h.paymentService.GetUser(ctx, customer.TelegramID)
	var info strings.Builder
	if user != nil {
		if user.ExpireAt.After(time.Now()) {
			info.WriteString(h.translation.GetText(lang, "subscription_active_hint"))
		} else {
			info.WriteString(h.translation.GetText(lang, "subscription_inactive_hint"))
		}
		info.WriteString("\n\n")
		expire := user.ExpireAt.Format("02.01.2006 15:04")
		status := "ACTIVE"
		if user.Status.Set {
			status = string(user.Status.Value)
		}
		lastClient := "-"
		if !user.LastConnectedNode.Null {
			lastClient = user.LastConnectedNode.Value.GetNodeName()
		}
		start := time.Now().Truncate(24 * time.Hour)
		usage, _ := h.paymentService.GetUserDailyUsage(ctx, user.UUID.String(), start, time.Now())
		limit := 0.0
		if v, ok := user.TrafficLimitBytes.Get(); ok {
			limit = float64(v)
		}
		info.WriteString(h.translation.GetText(lang, "account_info_header"))
		info.WriteString(fmt.Sprintf(h.translation.GetText(lang, "account_info_balance"), customer.Balance))
		info.WriteString(fmt.Sprintf(h.translation.GetText(lang, "account_info_expire"), expire))
		info.WriteString(fmt.Sprintf(h.translation.GetText(lang, "account_info_status"), status))
		info.WriteString(fmt.Sprintf(h.translation.GetText(lang, "account_info_last_client"), lastClient))
		info.WriteString(h.translation.GetText(lang, "traffic_info_header"))
		info.WriteString(fmt.Sprintf(h.translation.GetText(lang, "traffic_limit"), utils.FormatGB(usage), utils.FormatGB(limit)))
		info.WriteString(fmt.Sprintf(h.translation.GetText(lang, "traffic_total_used"), utils.FormatGB(user.LifetimeUsedTrafficBytes)))
		untilReset := time.Until(start.Add(24 * time.Hour))
		info.WriteString(fmt.Sprintf(h.translation.GetText(lang, "traffic_time_to_reset"), untilReset.Truncate(time.Second)))

	} else {
		info.WriteString(fmt.Sprintf(h.translation.GetText(lang, "balance_info"), int(customer.Balance)))
	}
	return info.String()
}

func (h *Handler) MenuCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	ctxWithTime, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	lang := update.Message.From.LanguageCode
	customer, err := h.findOrCreateCustomer(ctxWithTime, update.Message.Chat.ID, lang)
	if err != nil {
		slog.Error("find or create customer", "err", err)
		return
	}

	kb := h.buildStartKeyboard(customer, lang)
	text := fmt.Sprintf(h.translation.GetText(lang, "account_menu_text"), update.Message.From.FirstName) + "\n\n" + h.buildAccountInfo(ctxWithTime, customer, lang)

	m, err := b.SendMessage(ctxWithTime, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "🧹",
		ReplyMarkup: models.ReplyKeyboardRemove{RemoveKeyboard: true},
	})
	if err == nil {
		_, _ = b.DeleteMessage(ctxWithTime, &bot.DeleteMessageParams{ChatID: update.Message.Chat.ID, MessageID: m.ID})
	}

	_, err = b.SendMessage(ctxWithTime, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
		Text:        text,
	})
	if err != nil {
		slog.Error("Error sending /menu message", "err", err)
	}
}

func (h *Handler) HelpCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.ConnectCommandHandler(ctx, b, update)
}

func (h *Handler) PromoCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	lang := update.Message.From.LanguageCode
	customer, err := h.findOrCreateCustomer(ctx, update.Message.Chat.ID, lang)
	if err != nil {
		slog.Error("find or create customer", "err", err)
		return
	}

	parts := strings.Fields(update.Message.Text)
	if len(parts) > 1 {
		code := parts[1]
		if err := h.paymentService.ApplyPromocode(ctx, customer, code); err != nil {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: h.translation.GetText(lang, "promo_invalid")})
			return
		}
		until := ""
		if customer.ExpireAt != nil {
			until = customer.ExpireAt.Format("02.01.2006 15:04")
		}
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: fmt.Sprintf(h.translation.GetText(lang, "promo_applied"), until)})
		return
	}

	h.expectPromo(update.Message.Chat.ID)
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: h.translation.GetText(lang, "enter_promocode_prompt")})
}
