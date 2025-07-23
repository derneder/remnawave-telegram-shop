package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnawave-tg-shop-bot/internal/pkg/config"
)

func (h *Handler) OtherCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	lang := update.CallbackQuery.From.LanguageCode
	kb := [][]models.InlineKeyboardButton{
		{
			{Text: h.translation.GetText(lang, "faq_button"), CallbackData: CallbackFAQ},
			{Text: h.translation.GetText(lang, "traffic_limit_button"), CallbackData: CallbackTrafficLimit},
		},
		{
			{Text: h.translation.GetText(lang, "keys_button"), CallbackData: CallbackKeys},
			{Text: h.translation.GetText(lang, "qr_button"), CallbackData: CallbackQR},
		},
		{{Text: h.translation.GetText(lang, "short_button"), CallbackData: CallbackShortLink}},
		{{Text: h.translation.GetText(lang, "locations_button"), CallbackData: CallbackLocations}},
		{{Text: h.translation.GetText(lang, "regen_key_button"), CallbackData: CallbackRegenKey}},
		{{Text: h.translation.GetText(lang, "proxy_button"), CallbackData: CallbackProxy}},
	}
	if config.ServerStatusURL() != "" {
		kb = append(kb, []models.InlineKeyboardButton{{Text: h.translation.GetText(lang, "server_status_button"), URL: config.ServerStatusURL()}})
	}
	kb = append(kb, []models.InlineKeyboardButton{{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackStart}})

	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ParseMode:   models.ParseModeHTML,
		Text:        h.translation.GetText(lang, "other_menu_text"),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "there is no text in the message to edit") {
			slog.Error("send other menu", "err", err)
			return
		}
	} else {
		return
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		ParseMode:   models.ParseModeHTML,
		Text:        h.translation.GetText(lang, "other_menu_text"),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})

	if err != nil {
		slog.Error("send new other menu", "err", err)
	}
}

func (h *Handler) FAQCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.simpleBack(ctx, b, update, h.translation.GetText(update.CallbackQuery.From.LanguageCode, "coming_soon_text"))
}

func (h *Handler) TrafficLimitCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.simpleBack(ctx, b, update, h.translation.GetText(update.CallbackQuery.From.LanguageCode, "coming_soon_text"))
}

func (h *Handler) LocationsCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.simpleBack(ctx, b, update, h.translation.GetText(update.CallbackQuery.From.LanguageCode, "coming_soon_text"))
}

func (h *Handler) RegenKeyCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.simpleBack(ctx, b, update, h.translation.GetText(update.CallbackQuery.From.LanguageCode, "coming_soon_text"))
}

func (h *Handler) simpleBack(ctx context.Context, b *bot.Bot, update *models.Update, text string) {
	lang := update.CallbackQuery.From.LanguageCode
	kb := [][]models.InlineKeyboardButton{{{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackOther}}}
	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ParseMode:   models.ParseModeHTML,
		Text:        text,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("send simple back", "err", err)
	}
}

func (h *Handler) KeysCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	lang := update.CallbackQuery.From.LanguageCode
	customer, err := h.findOrCreateCustomer(ctx, update.CallbackQuery.From.ID, lang)
	if err != nil || customer.SubscriptionLink == nil {
		slog.Error("find customer", "err", err)
		return
	}
	resp, err := http.Get(*customer.SubscriptionLink)
	if err != nil {
		slog.Error("download keys", "err", err)
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	kb := [][]models.InlineKeyboardButton{{{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackOther}}}
	_, err = b.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		Document:    &models.InputFileUpload{Filename: "keys.txt", Data: bytes.NewReader(data)},
		Caption:     h.translation.GetText(lang, "keys_text"),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("send keys", "err", err)
	}
}

func (h *Handler) QRCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	lang := update.CallbackQuery.From.LanguageCode
	customer, err := h.findOrCreateCustomer(ctx, update.CallbackQuery.From.ID, lang)
	if err != nil || customer.SubscriptionLink == nil {
		slog.Error("find customer", "err", err)
		return
	}
	encoded := url.QueryEscape(*customer.SubscriptionLink)
	qrURL := "https://api.qrserver.com/v1/create-qr-code/?size=400x400&data=" + encoded
	resp, err := http.Get(qrURL)
	if err != nil {
		slog.Error("fetch qr", "err", err)
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	kb := [][]models.InlineKeyboardButton{{{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackOther}}}
	_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		Photo:       &models.InputFileUpload{Filename: "qr.png", Data: bytes.NewReader(data)},
		Caption:     fmt.Sprintf(h.translation.GetText(lang, "qr_text"), *customer.SubscriptionLink),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("send qr", "err", err)
	}
}

func (h *Handler) ShortLinkCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	lang := update.CallbackQuery.From.LanguageCode
	customer, err := h.findOrCreateCustomer(ctx, update.CallbackQuery.From.ID, lang)
	if err != nil || customer.SubscriptionLink == nil {
		slog.Error("find customer", "err", err)
		return
	}
	api := "https://tinyurl.com/api-create.php?url=" + url.QueryEscape(*customer.SubscriptionLink)
	resp, err := http.Get(api)
	if err != nil || resp.StatusCode >= 400 {
		alt := "https://is.gd/create.php?format=simple&url=" + url.QueryEscape(*customer.SubscriptionLink)
		resp, err = http.Get(alt)
		if err != nil {
			slog.Error("shorten", "err", err)
			return
		}
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	shortURL := string(data)
	h.shortLinks[customer.TelegramID] = append(h.shortLinks[customer.TelegramID], ShortLink{URL: shortURL, CreatedAt: time.Now()})
	kb := [][]models.InlineKeyboardButton{
		{{Text: h.translation.GetText(lang, "open_short_link_button"), URL: shortURL}},
		{{Text: h.translation.GetText(lang, "short_list_button"), CallbackData: CallbackShortList}},
		{{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackOther}},
	}
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ParseMode:   models.ParseModeHTML,
		Text:        fmt.Sprintf(h.translation.GetText(lang, "short_created_text"), shortURL),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("send short", "err", err)
	}
}

func (h *Handler) ShortListCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	lang := update.CallbackQuery.From.LanguageCode
	list := h.shortLinks[update.CallbackQuery.From.ID]
	var text string
	if len(list) == 0 {
		text = h.translation.GetText(lang, "short_list_text")
		text = fmt.Sprintf(text, "-")
	} else {
		var bld strings.Builder
		for i, l := range list {
			status := "Активна"
			if time.Since(l.CreatedAt) > 5*time.Minute {
				status = "Истекла"
			}
			fmt.Fprintf(&bld, "%d. %s\n   – %s\n", i+1, l.URL, status)
		}
		text = fmt.Sprintf(h.translation.GetText(lang, "short_list_text"), bld.String())
	}
	kb := [][]models.InlineKeyboardButton{{{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackOther}}}
	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ParseMode:   models.ParseModeHTML,
		Text:        text,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("send short list", "err", err)
	}
}

func (h *Handler) ProxyCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	lang := update.CallbackQuery.From.LanguageCode
	kb := [][]models.InlineKeyboardButton{}
	if config.TelegramProxyURL() != "" {
		kb = append(kb, []models.InlineKeyboardButton{{Text: h.translation.GetText(lang, "proxy_button"), URL: config.TelegramProxyURL()}})
	}
	kb = append(kb, []models.InlineKeyboardButton{{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackOther}})
	text := fmt.Sprintf(
		h.translation.GetText(lang, "proxy_details_text"),
		config.TelegramProxyChannel(),
		config.TelegramProxyChannel(),
		config.TelegramProxyHost(),
		config.TelegramProxyPort(),
		config.TelegramProxyKey(),
	)
	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ParseMode:   models.ParseModeHTML,
		Text:        text,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb},
	})
	if err != nil {
		slog.Error("send proxy", "err", err)
	}
}
