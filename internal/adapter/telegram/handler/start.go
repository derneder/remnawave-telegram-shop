package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/utils"
	"remnawave-tg-shop-bot/internal/service/payment"
	"remnawave-tg-shop-bot/internal/ui/menu"
)

func (h *Handler) StartCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	ctxWithTime, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	parts := strings.SplitN(update.Message.Text, " ", 2)
	if len(parts) == 1 {
		parts = strings.SplitN(update.Message.Text, "=", 2)
	}
	if len(parts) == 2 && strings.HasPrefix(parts[1], "ref_") {
		h.handleReferralStart(ctxWithTime, b, update, strings.TrimPrefix(parts[1], "ref_"))
		return
	}

	h.handlePlainStart(ctxWithTime, b, update)
}

func (h *Handler) handlePlainStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	langCode := update.Message.From.LanguageCode
	existingCustomer, err := h.customerRepository.FindByTelegramId(ctx, update.Message.Chat.ID)
	if err != nil {
		slog.Error("error finding customer by telegram id", "err", err)
		return
	}

	if existingCustomer == nil {
		_, err = h.customerRepository.Create(ctx, &domaincustomer.Customer{
			TelegramID: update.Message.Chat.ID,
			Language:   langCode,
			Balance:    0,
		})
		if err != nil {
			slog.Error("error creating customer", "err", err)
			return
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

	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		ReplyMarkup: models.ReplyKeyboardRemove{RemoveKeyboard: true},
	}); err != nil {
		slog.Error("send remove keyboard", "err", err)
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

func (h *Handler) handleReferralStart(ctx context.Context, b *bot.Bot, update *models.Update, payload string) {
	langCode := update.Message.From.LanguageCode
	referrerID, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		slog.Error("parse referrer id", "err", err)
		h.handlePlainStart(ctx, b, update)
		return
	}

	customer, err := h.customerRepository.FindByTelegramId(ctx, update.Message.Chat.ID)
	if err != nil {
		slog.Error("find customer", "err", err)
		return
	}

	if customer == nil {
		customer, err = h.customerRepository.Create(ctx, &domaincustomer.Customer{
			TelegramID: update.Message.Chat.ID,
			Language:   langCode,
			Balance:    0,
		})
		if err != nil {
			slog.Error("create customer", "err", err)
			return
		}
	}

	if referrerID == customer.TelegramID {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: h.translation.GetText(langCode, "ref.msg.self_ref")})
	} else if existing, _ := h.referralRepository.FindByReferee(ctx, customer.TelegramID); existing != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: h.translation.GetText(langCode, "ref.msg.already_registered")})
	} else {
		if err := h.referralRepository.Create(ctx, referrerID, customer.TelegramID); err == nil {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: h.translation.GetText(langCode, "ref.msg.saved")})
		}
	}

	text, kb := h.buildReferralInfo(customer, langCode)
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, ParseMode: models.ParseModeHTML, Text: text, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}})
	if err != nil {
		slog.Error("send referral info", "err", err)
	}
}

func (h *Handler) buildReferralInfo(customer *domaincustomer.Customer, lang string) (string, [][]models.InlineKeyboardButton) {
	botURL := strings.TrimPrefix(config.BotURL(), "https://t.me/")
	botURL = strings.TrimPrefix(botURL, "http://t.me/")
	refLink := menu.BuildReferralLink(botURL, fmt.Sprintf("ref_%d", customer.TelegramID))

	invited := 0
	subscribed := 0
	bonusTotal := 0

	text := fmt.Sprintf(h.translation.GetText(lang, "ref.msg.full"), invited, subscribed, bonusTotal, refLink, config.GetReferralBonus(), config.GetReferralBonus())

	kb := [][]models.InlineKeyboardButton{{{Text: h.translation.GetText(lang, "ref.button.invite"), URL: refLink}}}
	return text, kb
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

	inlineKeyboard := menu.BuildLKMenu(langCode, existingCustomer, isAdmin(existingCustomer.TelegramID))

	text := fmt.Sprintf(h.translation.GetText(langCode, "account_menu_text"), callback.From.FirstName) + "\n\n" + h.buildAccountInfo(ctxWithTime, existingCustomer, langCode)

	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctxWithTime, b, curMsg, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msgID,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboard},
		Text:        text,
	})
	if err != nil {
		slog.Error("Error sending /start message", "err", err)
	}
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

	kb := menu.BuildLKMenu(lang, customer, isAdmin(customer.TelegramID))
	text := fmt.Sprintf(h.translation.GetText(lang, "account_menu_text"), update.Message.From.FirstName) + "\n\n" + h.buildAccountInfo(ctxWithTime, customer, lang)

	if _, err := b.SendMessage(ctxWithTime, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		ReplyMarkup: models.ReplyKeyboardRemove{RemoveKeyboard: true},
	}); err != nil {
		slog.Error("send remove keyboard", "err", err)
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
		promo, err := h.paymentService.ApplyPromocode(ctx, customer, code)
		if err != nil {
			var text string
			if errors.Is(err, payment.ErrPromocodeNotFound) {
				text = h.translation.GetText(lang, "promo_not_found")
			} else if errors.Is(err, payment.ErrPromocodeExpired) {
				text = h.translation.GetText(lang, "promo_expired")
			} else if errors.Is(err, payment.ErrPromocodeLimitExced) {
				text = h.translation.GetText(lang, "promo_limit_reached")
			} else {
				text = h.translation.GetText(lang, "promo_invalid")
			}
			if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: text}); serr != nil {
				slog.Error("send promo invalid", "err", serr)
			}
			return
		}
		if promo.Type == 2 {
			if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, ParseMode: models.ParseModeHTML, Text: fmt.Sprintf(h.translation.GetText(lang, "promo_balance_applied"), promo.Amount/100, int(customer.Balance))}); serr != nil {
				slog.Error("send balance promo", "err", serr)
			}
			return
		}
		until := ""
		if customer.ExpireAt != nil {
			until = customer.ExpireAt.Format("02.01.2006 15:04")
		}
		if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: fmt.Sprintf(h.translation.GetText(lang, "promo_applied"), until)}); serr != nil {
			slog.Error("send promo applied", "err", serr)
		}
		return
	}

	h.expectPromo(update.Message.Chat.ID)
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: h.translation.GetText(lang, "promo.activate.prompt")}); err != nil {
		slog.Error("send promo activate prompt", "err", err)
	}
}
