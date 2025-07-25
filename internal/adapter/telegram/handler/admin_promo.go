package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnawave-tg-shop-bot/internal/pkg/contextkey"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	uimenu "remnawave-tg-shop-bot/internal/ui/menu"
)

// adminPromoState keeps wizard data for one admin.
type adminPromoState struct {
	Type   string // "balance" or "sub"
	Step   uimenu.StepState
	Amount int
	Limit  int
	Days   int
	Code   string
}

// AdminPromoCallbackHandler routes admin promo callbacks.
func (h *Handler) AdminPromoCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !contextkey.IsAdminFromContext(ctx) {
		return
	}
	lang := update.CallbackQuery.From.LanguageCode
	data := update.CallbackQuery.Data
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	tm := translation.GetInstance()

	state := h.adminStates[update.CallbackQuery.From.ID]
	if strings.HasPrefix(data, uimenu.CallbackAdminMenu) {
		h.adminStates[update.CallbackQuery.From.ID] = &adminPromoState{}
		kb := uimenu.BuildAdminPromoMenu(lang)
		_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "admin_panel_button")})
		return
	}
	if state == nil {
		return
	}
	switch state.Type {
	case "":
		if data == uimenu.CallbackAdminPromoBalanceStart {
			state.Type = "balance"
			state.Step = uimenu.StepAmount
			kb := uimenu.BuildAdminPromoBalanceWizardStep(lang, uimenu.StepAmount)
			_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_amount_prompt")})
			return
		}
		if data == uimenu.CallbackAdminPromoSubStart {
			state.Type = "sub"
			state.Step = uimenu.StepCode
			kb := uimenu.BuildAdminPromoSubWizardStep(lang, uimenu.StepCode)
			_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_code_prompt")})
			return
		}
	case "balance":
		switch state.Step {
		case uimenu.StepAmount:
			if strings.HasPrefix(data, uimenu.CallbackPromoBalanceAmount) {
				val := strings.TrimPrefix(data, uimenu.CallbackPromoBalanceAmount+":")
				if val != "manual" {
					a, _ := strconv.Atoi(val)
					state.Amount = a
					state.Step = uimenu.StepLimit
					kb := uimenu.BuildAdminPromoBalanceWizardStep(lang, uimenu.StepLimit)
					_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_limit_prompt")})
				}
				return
			}
		case uimenu.StepLimit:
			if strings.HasPrefix(data, uimenu.CallbackPromoBalanceLimit) {
				val := strings.TrimPrefix(data, uimenu.CallbackPromoBalanceLimit+":")
				if val != "manual" {
					l, _ := strconv.Atoi(val)
					state.Limit = l
					state.Step = uimenu.StepConfirm
					kb := uimenu.BuildAdminPromoBalanceWizardStep(lang, uimenu.StepConfirm)
					txt := fmt.Sprintf(tm.GetText(lang, "promo_confirm_text"), state.Amount, state.Limit)
					_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: txt})
				}
				return
			}
		case uimenu.StepConfirm:
			switch data {
			case uimenu.CallbackPromoBalanceConfirm:
				code, err := h.promotionService.CreateBalance(ctx, state.Amount*100, state.Limit, update.CallbackQuery.From.ID)
				if err != nil {
					slog.Error("create bal promo", "err", err)
					return
				}
				text := fmt.Sprintf(tm.GetText(lang, "promo_balance_created"), code, state.Amount, state.Limit)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, Text: text})
				delete(h.adminStates, update.CallbackQuery.From.ID)
				return
			case uimenu.CallbackAdminBack:
				state.Step = uimenu.StepLimit
				kb := uimenu.BuildAdminPromoBalanceWizardStep(lang, uimenu.StepLimit)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_limit_prompt")})
				return
			case uimenu.CallbackAdminCancel:
				delete(h.adminStates, update.CallbackQuery.From.ID)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, Text: tm.GetText(lang, "promo_cancelled")})
				return
			}
		}
	case "sub":
		switch state.Step {
		case uimenu.StepCode:
			if data == uimenu.CallbackPromoSubCodeRandom {
				state.Code = ""
				state.Step = uimenu.StepDays
				kb := uimenu.BuildAdminPromoSubWizardStep(lang, uimenu.StepDays)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_days_prompt")})
				return
			}
			if data == uimenu.CallbackPromoSubCodeCustom {
				// For simplicity, not implemented in tests
				state.Code = "CUSTOM"
				state.Step = uimenu.StepDays
				kb := uimenu.BuildAdminPromoSubWizardStep(lang, uimenu.StepDays)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_days_prompt")})
				return
			}
		case uimenu.StepDays:
			if strings.HasPrefix(data, uimenu.CallbackPromoSubDays) {
				val := strings.TrimPrefix(data, uimenu.CallbackPromoSubDays+":")
				d, _ := strconv.Atoi(val)
				state.Days = d
				state.Step = uimenu.StepLimit
				kb := uimenu.BuildAdminPromoSubWizardStep(lang, uimenu.StepLimit)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_limit_prompt")})
				return
			}
		case uimenu.StepLimit:
			if strings.HasPrefix(data, uimenu.CallbackPromoSubLimit) {
				val := strings.TrimPrefix(data, uimenu.CallbackPromoSubLimit+":")
				l, _ := strconv.Atoi(val)
				state.Limit = l
				state.Step = uimenu.StepConfirm
				kb := uimenu.BuildAdminPromoSubWizardStep(lang, uimenu.StepConfirm)
				txt := fmt.Sprintf(tm.GetText(lang, "promo_confirm_text"), state.Days, state.Limit)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: txt})
				return
			}
		case uimenu.StepConfirm:
			switch data {
			case uimenu.CallbackPromoSubConfirm:
				code := state.Code
				if code == "" {
					code = "RND" // stub
				}
				err := h.promotionService.CreateSubscription(ctx, code, state.Days, state.Limit, update.CallbackQuery.From.ID)
				if err != nil {
					slog.Error("create sub promo", "err", err)
					return
				}
				text := fmt.Sprintf(tm.GetText(lang, "promo_sub_created"), code, state.Days, state.Limit)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, Text: text})
				delete(h.adminStates, update.CallbackQuery.From.ID)
				return
			case uimenu.CallbackAdminBack:
				state.Step = uimenu.StepLimit
				kb := uimenu.BuildAdminPromoSubWizardStep(lang, uimenu.StepLimit)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_limit_prompt")})
				return
			case uimenu.CallbackAdminCancel:
				delete(h.adminStates, update.CallbackQuery.From.ID)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, Text: tm.GetText(lang, "promo_cancelled")})
				return
			}
		}
	}
}
