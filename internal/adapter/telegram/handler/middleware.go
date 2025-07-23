package handler

import (
	"context"

	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
)

func (h *Handler) CreateCustomerIfNotExistMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		var telegramId int64
		var langCode string
		if update.Message != nil {
			telegramId = update.Message.From.ID
			langCode = update.Message.From.LanguageCode
		} else if update.CallbackQuery != nil {
			telegramId = update.CallbackQuery.From.ID
			langCode = update.CallbackQuery.From.LanguageCode
		}
		existingCustomer, err := h.customerRepository.FindByTelegramId(ctx, telegramId)
		if err != nil {
			slog.Error("error finding customer by telegram id", "err", err)
			return
		}

		if existingCustomer == nil {
			_, err = h.customerRepository.Create(ctx, &domaincustomer.Customer{
				TelegramID: telegramId,
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

		next(ctx, b, update)
	}
}
