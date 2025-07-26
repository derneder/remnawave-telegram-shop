package handler

import (
	"context"
	"fmt"
	"strings"

	"log/slog"
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func parseCallbackData(data string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(data, "?")
	if len(parts) < 2 {
		return result
	}
	params := strings.Split(parts[1], "&")
	for _, param := range params {
		kv := strings.SplitN(param, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

// callbackChatMessage extracts chat and message IDs from a callback update.
// It supports both accessible and inaccessible messages.
func callbackChatMessage(upd *models.Update) (int64, int, bool) {
	if upd == nil || upd.CallbackQuery == nil {
		return 0, 0, false
	}
	m := upd.CallbackQuery.Message
	if m.Message != nil {
		return m.Message.Chat.ID, m.Message.ID, true
	}
	if m.InaccessibleMessage != nil {
		return m.InaccessibleMessage.Chat.ID, m.InaccessibleMessage.MessageID, true
	}
	return 0, 0, false
}

// getCallbackIDs wraps callbackChatMessage and returns an error when
// the message is not available.
func getCallbackIDs(upd *models.Update) (int64, int, error) {
	chatID, msgID, ok := callbackChatMessage(upd)
	if !ok {
		return 0, 0, fmt.Errorf("callback message missing")
	}
	return chatID, msgID, nil
}

func (h *Handler) findOrCreateCustomer(ctx context.Context, telegramID int64, lang string) (*domaincustomer.Customer, error) {
	customer, err := h.customerRepository.FindByTelegramId(ctx, telegramID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		customer, err = h.customerRepository.Create(ctx, &domaincustomer.Customer{TelegramID: telegramID, Language: lang, Balance: 0})
		if err != nil {
			slog.Error("create customer", "err", err)
			return nil, err
		}
	}
	return customer, nil
}
func (h *Handler) buildPaymentBackData(month int, amount int) string {
	if month == 0 {
		return fmt.Sprintf("%s?amount=%d", CallbackTopupMethod, amount)
	}
	return fmt.Sprintf("%s?month=%d&amount=%d", CallbackSell, month, amount)
}

// callbackInfo returns chat id, message id and actual message pointer from a callback update.
func callbackInfo(upd *models.Update) (int64, int, *models.Message, error) {
	chatID, msgID, err := getCallbackIDs(upd)
	if err != nil {
		return 0, 0, nil, err
	}
	var curMsg *models.Message
	if upd.CallbackQuery.Message.Message != nil {
		curMsg = upd.CallbackQuery.Message.Message
	}
	return chatID, msgID, curMsg, nil
}

// editCallback wraps SafeEditMessageText using IDs extracted from update.
func editCallback(ctx context.Context, b *bot.Bot, upd *models.Update, params *bot.EditMessageTextParams) error {
	chatID, msgID, curMsg, err := callbackInfo(upd)
	if err != nil {
		return err
	}
	params.ChatID = chatID
	params.MessageID = msgID
	_, err = SafeEditMessageText(ctx, b, curMsg, params)
	return err
}

func editCallbackWithLog(ctx context.Context, b *bot.Bot, upd *models.Update, params *bot.EditMessageTextParams, logMsg string) {
	if err := editCallback(ctx, b, upd, params); err != nil {
		slog.Error(logMsg, "err", err)
	}
}
