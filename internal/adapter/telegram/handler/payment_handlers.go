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
	pg "remnawave-tg-shop-bot/internal/repository/pg"

	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/contextkey"
)

func (h *Handler) BuyCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	langCode := update.CallbackQuery.From.LanguageCode

	var priceButtons []models.InlineKeyboardButton

	if config.Price1() > 0 {
		priceButtons = append(priceButtons, models.InlineKeyboardButton{
			Text:         h.translation.GetText(langCode, "month_1"),
			CallbackData: fmt.Sprintf("%s?month=%d&amount=%d", CallbackSell, 1, config.Price1()),
		})
	}

	if config.Price3() > 0 {
		priceButtons = append(priceButtons, models.InlineKeyboardButton{
			Text:         h.translation.GetText(langCode, "month_3"),
			CallbackData: fmt.Sprintf("%s?month=%d&amount=%d", CallbackSell, 3, config.Price3()),
		})
	}

	if config.Price6() > 0 {
		priceButtons = append(priceButtons, models.InlineKeyboardButton{
			Text:         h.translation.GetText(langCode, "month_6"),
			CallbackData: fmt.Sprintf("%s?month=%d&amount=%d", CallbackSell, 6, config.Price6()),
		})
	}

	keyboard := [][]models.InlineKeyboardButton{}

	if len(priceButtons) == 3 {
		keyboard = append(keyboard, priceButtons[:2])
		keyboard = append(keyboard, priceButtons[2:])
	} else if len(priceButtons) > 0 {
		keyboard = append(keyboard, priceButtons)
	}

	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{Text: h.translation.GetText(langCode, "back_to_account_button"), CallbackData: CallbackStart},
	})

	customer, _ := h.customerRepository.FindByTelegramId(ctx, chatID)
	bal := 0
	if customer != nil {
		bal = int(customer.Balance)
	}
	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	var lines []string

	if config.Price1() > 0 {
		lines = append(lines, fmt.Sprintf(
			h.translation.GetText(langCode, "plan_line"),
			"✨",
			h.translation.GetText(langCode, "month_1"),
			config.Price1(),
		))
	}

	if config.Price3() > 0 {
		lines = append(lines, fmt.Sprintf(
			h.translation.GetText(langCode, "plan_line"),
			"❤️‍🔥",
			h.translation.GetText(langCode, "month_3"),
			config.Price3(),
		))
	}

	if config.Price6() > 0 {
		lines = append(lines, fmt.Sprintf(
			h.translation.GetText(langCode, "plan_line"),
			"🔥",
			h.translation.GetText(langCode, "month_6"),
			config.Price6(),
		))
	}

	_, err := SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: msgID,
		ParseMode: models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
		Text: fmt.Sprintf(h.translation.GetText(langCode, "choose_plan_text"), bal, strings.Join(lines, "\n")),
	})

	if err != nil {
		slog.Error("Error sending buy message", "err", err)
	}
}

func (h *Handler) SellCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	callbackQuery := parseCallbackData(update.CallbackQuery.Data)
	langCode := update.CallbackQuery.From.LanguageCode
	month := callbackQuery["month"]

	keyboard := [][]models.InlineKeyboardButton{
		{{Text: h.translation.GetText(langCode, "buy_sub_balance_button"), CallbackData: fmt.Sprintf("%s?month=%s", CallbackPayFromBal, month)}},
	}

	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{Text: h.translation.GetText(langCode, "back_button"), CallbackData: CallbackBuy},
	})

	customer, _ := h.customerRepository.FindByTelegramId(ctx, chatID)
	bal := 0
	if customer != nil {
		bal = int(customer.Balance)
	}
	var (
		line      string
		price     int
		emoji     string
		monthText string
	)

	switch month {
	case "1":
		price = config.Price1()
		emoji = "✨"
		monthText = h.translation.GetText(langCode, "month_1")
	case "3":
		price = config.Price3()
		emoji = "❤️‍🔥"
		monthText = h.translation.GetText(langCode, "month_3")
	case "6":
		price = config.Price6()
		emoji = "🔥"
		monthText = h.translation.GetText(langCode, "month_6")
	}

	line = fmt.Sprintf(
		h.translation.GetText(langCode, "plan_line"),
		emoji,
		monthText,
		price,
	)

	text := fmt.Sprintf(h.translation.GetText(langCode, "choose_plan_text"), bal, line)

	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: msgID,
		ParseMode: models.ParseModeHTML,
		Text:      text,
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	})

	if err != nil {
		slog.Error("Error sending sell message", "err", err)
	}
}

func (h *Handler) PaymentCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	callbackQuery := parseCallbackData(update.CallbackQuery.Data)
	month, err := strconv.Atoi(callbackQuery["month"])
	if err != nil {
		slog.Error("Error getting month from query", "err", err)
		return
	}

	invoiceType := pg.InvoiceType(callbackQuery["invoiceType"])
	amountParam, _ := strconv.Atoi(callbackQuery["amount"])

	var price int
	if month == 0 {
		price = amountParam
	} else if invoiceType == pg.InvoiceTypeTelegram {
		price = config.StarsPrice(month)
	} else {
		price = config.Price(month)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	customer, err := h.customerRepository.FindByTelegramId(ctx, chatID)
	if err != nil {
		slog.Error("Error finding customer", "err", err)
		return
	}
	if customer == nil {
		slog.Error("customer not exist", "chatID", chatID, "err", err)
		return
	}

	ctxWithUsername := context.WithValue(ctx, contextkey.Username, contextkey.CleanUsername(update.CallbackQuery.From.Username))
	paymentURL, purchaseId, err := h.paymentService.CreatePurchase(ctxWithUsername, price, month, customer, invoiceType)
	if err != nil {
		slog.Error("Error creating payment", "err", err)
		return
	}

	langCode := update.CallbackQuery.From.LanguageCode

	message, err := b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:    chatID,
		MessageID: msgID,
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: h.translation.GetText(langCode, "pay_button"), URL: paymentURL},
					{Text: h.translation.GetText(langCode, "back_button"), CallbackData: h.buildPaymentBackData(month, amountParam)},
				},
			},
		},
	})
	if err != nil {
		slog.Error("Error updating sell message", "err", err)
		return
	}
	h.cache.Set(purchaseId, message.ID)
}

func (h *Handler) PreCheckoutCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.AnswerPreCheckoutQuery(ctx, &bot.AnswerPreCheckoutQueryParams{
		PreCheckoutQueryID: update.PreCheckoutQuery.ID,
		OK:                 true,
	})
	if err != nil {
		slog.Error("Error sending answer pre checkout query", "err", err)
	}
}

func (h *Handler) SuccessPaymentHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	payload := strings.Split(update.Message.SuccessfulPayment.InvoicePayload, "&")
	purchaseId, err := strconv.Atoi(payload[0])
	username := payload[1]
	if err != nil {
		slog.Error("Error parsing purchase id", "err", err)
		return
	}

	ctxWithUsername := context.WithValue(ctx, contextkey.Username, contextkey.CleanUsername(username))
	err = h.paymentService.ProcessPurchaseById(ctxWithUsername, int64(purchaseId))
	if err != nil {
		slog.Error("Error processing purchase", "err", err)
	}
}

func (h *Handler) BalanceCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	lang := update.CallbackQuery.From.LanguageCode
	customer, _ := h.customerRepository.FindByTelegramId(ctx, chatID)
	if customer == nil {
		return
	}

	text := fmt.Sprintf(h.translation.GetText(lang, "balance_menu_text"), int(customer.Balance))

	keyboard := [][]models.InlineKeyboardButton{
		{{Text: h.translation.GetText(lang, "topup_button"), CallbackData: CallbackTopup}},
		{{Text: h.translation.GetText(lang, "buy_sub_balance_button"), CallbackData: CallbackBuy}},
		{{Text: h.translation.GetText(lang, "back_to_account_button"), CallbackData: CallbackStart}},
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
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: keyboard},
	})
	if err != nil {
		slog.Error("Error sending balance message", "err", err)
	}
}

func (h *Handler) TopupCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	lang := update.CallbackQuery.From.LanguageCode
	customer, _ := h.customerRepository.FindByTelegramId(ctx, chatID)
	if customer == nil {
		return
	}
	keyboard := [][]models.InlineKeyboardButton{
		{
			{Text: "💵 100", CallbackData: fmt.Sprintf("%s?amount=100", CallbackTopupMethod)},
			{Text: "💵 200", CallbackData: fmt.Sprintf("%s?amount=200", CallbackTopupMethod)},
			{Text: "💵 300", CallbackData: fmt.Sprintf("%s?amount=300", CallbackTopupMethod)},
		},
		{
			{Text: "💵 500", CallbackData: fmt.Sprintf("%s?amount=500", CallbackTopupMethod)},
			{Text: "💵 750", CallbackData: fmt.Sprintf("%s?amount=750", CallbackTopupMethod)},
			{Text: "💵 1000", CallbackData: fmt.Sprintf("%s?amount=1000", CallbackTopupMethod)},
		},
		{
			{Text: "💵 1500", CallbackData: fmt.Sprintf("%s?amount=1500", CallbackTopupMethod)},
			{Text: "💵 3000", CallbackData: fmt.Sprintf("%s?amount=3000", CallbackTopupMethod)},
			{Text: "💵 5000", CallbackData: fmt.Sprintf("%s?amount=5000", CallbackTopupMethod)},
		},
		{{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackBalance}},
	}
	text := fmt.Sprintf(h.translation.GetText(lang, "topup_intro_text"), int(customer.Balance))
	params := &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msgID,
		Text:        text,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: keyboard},
	}
	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg, params)
	if err != nil {
		slog.Error("Error sending topup message", "err", err)
	}
}

func (h *Handler) TopupMethodCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	data := parseCallbackData(update.CallbackQuery.Data)
	amount := data["amount"]
	lang := update.CallbackQuery.From.LanguageCode

	var keyboard [][]models.InlineKeyboardButton
	for _, p := range h.paymentService.EnabledProviders() {
		switch p.Type() {
		case pg.InvoiceTypeCrypto:
			keyboard = append(keyboard, []models.InlineKeyboardButton{{Text: h.translation.GetText(lang, "crypto_button"), CallbackData: fmt.Sprintf("%s?month=0&invoiceType=%s&amount=%s", CallbackPayment, pg.InvoiceTypeCrypto, amount)}})
		case pg.InvoiceTypeTribute:
			keyboard = append(keyboard, []models.InlineKeyboardButton{{Text: h.translation.GetText(lang, "tribute_button"), CallbackData: fmt.Sprintf("%s?month=0&invoiceType=%s&amount=%s", CallbackPayment, pg.InvoiceTypeTribute, amount)}})
		}
	}
	if config.IsTelegramStarsEnabled() {
		keyboard = append(keyboard, []models.InlineKeyboardButton{{Text: h.translation.GetText(lang, "stars_button"), CallbackData: fmt.Sprintf("%s?month=0&invoiceType=%s&amount=%s", CallbackPayment, pg.InvoiceTypeTelegram, amount)}})
	}
	keyboard = append(keyboard, []models.InlineKeyboardButton{{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackTopup}})

	_, err = b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:      chatID,
		MessageID:   msgID,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: keyboard},
	})
	if err != nil {
		slog.Error("Error sending topup methods", "err", err)
	}
}
func (h *Handler) PayFromBalanceCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID, _, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	data := parseCallbackData(update.CallbackQuery.Data)
	month, _ := strconv.Atoi(data["month"])

	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	customer, err := h.customerRepository.FindByTelegramId(ctxTimeout, chatID)
	if err != nil || customer == nil {
		return
	}
	if err := h.paymentService.PurchaseFromBalance(ctxTimeout, customer, month); err != nil {
		slog.Error("error pay from balance", "err", err)
	}
}
