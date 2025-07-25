package handler

import (
	"context"
	"fmt"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	"remnawave-tg-shop-bot/utils"
)

func (h *Handler) ConnectCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	customer, err := h.customerRepository.FindByTelegramId(ctx, update.Message.Chat.ID)
	if err != nil {
		slog.Error("Error finding customer", "err", err)
		return
	}
	if customer == nil {
		slog.Error("customer not exist", "telegramId", utils.MaskHalfInt64(update.Message.Chat.ID), "error", err)
		return
	}

	langCode := update.Message.From.LanguageCode

	isDisabled := true
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      buildConnectText(customer, langCode),
		ParseMode: models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: &isDisabled,
		},
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: h.translation.GetText(langCode, "back_to_account_button"), CallbackData: CallbackStart}},
			},
		},
	})

	if err != nil {
		slog.Error("Error sending connect message", "err", err)
	}
}

func (h *Handler) ConnectCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID, msgID, ok := callbackChatMessage(update)
	if !ok {
		slog.Error("callback message missing")
		return
	}

	customer, err := h.customerRepository.FindByTelegramId(ctx, chatID)
	if err != nil {
		slog.Error("Error finding customer", "err", err)
		return
	}
	if customer == nil {
		slog.Error("customer not exist", "telegramId", utils.MaskHalfInt64(chatID), "error", err)
		return
	}

	langCode := update.CallbackQuery.From.LanguageCode

	var markup [][]models.InlineKeyboardButton
	if config.IsWepAppLinkEnabled() {
		if customer.SubscriptionLink != nil && customer.ExpireAt.After(time.Now()) {
			markup = append(markup, []models.InlineKeyboardButton{{Text: h.translation.GetText(langCode, "connect_button"),
				WebApp: &models.WebAppInfo{
					URL: *customer.SubscriptionLink,
				}}})
		}
	}
	markup = append(markup, []models.InlineKeyboardButton{{Text: h.translation.GetText(langCode, "back_to_account_button"), CallbackData: CallbackStart}})

	isDisabled := true
	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: msgID,
		ParseMode: models.ParseModeHTML,
		Text:      buildConnectText(customer, langCode),
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: &isDisabled,
		},
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: markup,
		},
	})

	if err != nil {
		slog.Error("Error sending connect message", "err", err)
	}
}

func buildConnectText(customer *domaincustomer.Customer, langCode string) string {
	var info strings.Builder

	tm := translation.GetInstance()

	if customer.ExpireAt != nil {
		currentTime := time.Now()

		if currentTime.Before(*customer.ExpireAt) {
			formattedDate := customer.ExpireAt.Format("02.01.2006 15:04")

			subscriptionActiveText := tm.GetText(langCode, "subscription_active")
			info.WriteString(fmt.Sprintf(subscriptionActiveText, formattedDate))

			if customer.SubscriptionLink != nil && *customer.SubscriptionLink != "" {
				if config.IsWepAppLinkEnabled() {
				} else {
					subscriptionLinkText := tm.GetText(langCode, "subscription_link")
					info.WriteString(fmt.Sprintf(subscriptionLinkText, *customer.SubscriptionLink))
				}
			}
		} else {
			noSubscriptionText := tm.GetText(langCode, "no_subscription")
			info.WriteString(noSubscriptionText)
		}
	} else {
		noSubscriptionText := tm.GetText(langCode, "no_subscription")
		info.WriteString(noSubscriptionText)
	}

	return info.String()
}
