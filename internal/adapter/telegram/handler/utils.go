package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"

	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/utils"
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

func (h *Handler) promptPromoMonths(ctx context.Context, b *bot.Bot, msg *models.Message, lang string, uses int, customer *domaincustomer.Customer) {
	price1 := config.Price1() * uses
	price3 := config.Price3() * uses
	price6 := config.Price6() * uses

	text1 := h.translation.GetText(lang, "month_1")
	text3 := h.translation.GetText(lang, "month_3")
	text6 := h.translation.GetText(lang, "month_6")

	if !config.IsAdmin(customer.TelegramID) {
		text1 = fmt.Sprintf("%s — %s ₽", text1, utils.FormatPrice(price1))
		text3 = fmt.Sprintf("%s — %s ₽", text3, utils.FormatPrice(price3))
		text6 = fmt.Sprintf("%s — %s ₽", text6, utils.FormatPrice(price6))
	}

	kb := [][]models.InlineKeyboardButton{
		{{Text: text1, CallbackData: fmt.Sprintf("%s?m=1&u=%d", CallbackPromoCreate, uses)}},
		{{Text: text3, CallbackData: fmt.Sprintf("%s?m=3&u=%d", CallbackPromoCreate, uses)}},
		{{Text: text6, CallbackData: fmt.Sprintf("%s?m=6&u=%d", CallbackPromoCreate, uses)}},
		{{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackReferral}},
	}

	bal := int(customer.Balance)
	if _, err := SafeEditMessageText(ctx, b, msg, &bot.EditMessageTextParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
		ParseMode: models.ParseModeHTML,
		Text:      fmt.Sprintf(h.translation.GetText(lang, "promo_choose_plan"), bal),
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: kb,
		},
	}); err != nil {
		slog.Error("send promo plan", "err", err)
	}
}

func (h *Handler) promptPromoUses(ctx context.Context, b *bot.Bot, msg *models.Message, lang string) {
	opts := []int{1, 10, 30, 70, 100}
	var kb [][]models.InlineKeyboardButton
	for _, u := range opts {
		word := h.translation.GetText(lang, "activation_plural")
		if u == 1 {
			word = h.translation.GetText(lang, "activation_singular")
		}
		label := fmt.Sprintf("%d %s", u, word)
		kb = append(kb, []models.InlineKeyboardButton{{Text: label, CallbackData: fmt.Sprintf("%s?u=%d", CallbackPromoCreate, u)}})
	}
	kb = append(kb, []models.InlineKeyboardButton{{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackReferral}})
	if _, err := SafeEditMessageText(ctx, b, msg, &bot.EditMessageTextParams{
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Text:        h.translation.GetText(lang, "promo_choose_uses"),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	}); err != nil {
		slog.Error("send promo uses", "err", err)
	}
}
