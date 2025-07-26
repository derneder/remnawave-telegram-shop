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
			referrer, err := h.customerRepository.FindByTelegramId(ctx, referrerId)
			if err == nil && referrer != nil {
				if err := h.referralRepository.Create(ctx, referrerId, existingCustomer.TelegramID); err == nil {
					bonus := float64(config.GetReferralBonus())
					_ = h.customerRepository.UpdateFields(ctx, referrer.ID, map[string]interface{}{"balance": referrer.Balance + bonus})
					_ = h.customerRepository.UpdateFields(ctx, existingCustomer.ID, map[string]interface{}{"balance": bonus})
					if _, err := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: referrer.TelegramID, Text: h.translation.GetText(referrer.Language, "referral_bonus_granted")}); err != nil {
						slog.Error("send referral bonus", "err", err)
					}
					if _, err := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: existingCustomer.TelegramID, Text: h.translation.GetText(langCode, "referral_bonus_granted")}); err != nil {
						slog.Error("send referral bonus", "err", err)
					}
					slog.Info("referral created", "referrerId", utils.MaskHalfInt64(referrerId), "refereeId", utils.MaskHalfInt64(existingCustomer.TelegramID))
				}
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
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: h.translation.GetText(lang, "enter_promocode_prompt")}); err != nil {
		slog.Error("send enter promocode prompt", "err", err)
	}
}
