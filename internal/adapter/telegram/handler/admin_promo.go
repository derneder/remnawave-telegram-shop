package handler

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnawave-tg-shop-bot/internal/pkg/contextkey"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	uimenu "remnawave-tg-shop-bot/internal/ui/menu"
)

var promoCodeRe = regexp.MustCompile(`^[A-Z0-9-]+$`)

// adminPromoState keeps wizard data for one admin.
type adminPromoState struct {
	Type   string // "balance" or "sub"
	Step   uimenu.StepState
	Amount int
	Limit  int
	Months int
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
	if strings.HasPrefix(data, uimenu.CallbackPromoAdminMenu) {
		h.adminStates[update.CallbackQuery.From.ID] = &adminPromoState{}
		h.clearAdminInputs(update.CallbackQuery.From.ID)
		kb := uimenu.BuildAdminPromoMenu(lang)
		_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "admin_panel_button")})
		return
	}
	if state == nil {
		return
	}
	switch state.Type {
	case "":
		if data == uimenu.CallbackPromoAdminBalanceStart {
			state.Type = "balance"
			state.Step = uimenu.StepAmount
			kb := uimenu.BuildAdminPromoBalanceWizardStep(lang, uimenu.StepAmount)
			_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_amount_prompt")})
			return
		}
		if data == uimenu.CallbackPromoAdminSubStart {
			state.Type = "sub"
			state.Step = uimenu.StepCode
			kb := uimenu.BuildAdminPromoSubWizardStep(lang, uimenu.StepCode)
			_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_code_prompt")})
			return
		}
	case "balance":
		switch state.Step {
		case uimenu.StepAmount:
			if strings.HasPrefix(data, uimenu.CallbackPromoAdminBalanceAmount) {
				val := strings.TrimPrefix(data, uimenu.CallbackPromoAdminBalanceAmount+":")
				if val == "manual" {
					h.expectAmount(update.CallbackQuery.From.ID)
					_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, Text: tm.GetText(lang, "promo.balance.amount.manual_prompt")})
				} else {
					a, _ := strconv.Atoi(val)
					state.Amount = a
					state.Step = uimenu.StepLimit
					kb := uimenu.BuildAdminPromoBalanceWizardStep(lang, uimenu.StepLimit)
					_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_limit_prompt")})
				}
				return
			}
		case uimenu.StepLimit:
			if strings.HasPrefix(data, uimenu.CallbackPromoAdminBalanceLimit) {
				val := strings.TrimPrefix(data, uimenu.CallbackPromoAdminBalanceLimit+":")
				if val == "manual" {
					h.expectLimit(update.CallbackQuery.From.ID)
					_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, Text: tm.GetText(lang, "promo.limit.manual_prompt")})
				} else {
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
			case uimenu.CallbackPromoAdminBalanceConfirm:
				code, err := h.promotionService.CreateBalance(ctx, state.Amount*100, state.Limit, update.CallbackQuery.From.ID)
				if err != nil {
					slog.Error("create bal promo", "err", err)
					return
				}
				text := fmt.Sprintf(tm.GetText(lang, "promo_balance_created"), code, state.Amount, state.Limit)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, Text: text})
				delete(h.adminStates, update.CallbackQuery.From.ID)
				h.clearAdminInputs(update.CallbackQuery.From.ID)
				return
			case uimenu.CallbackPromoAdminBack:
				state.Step = uimenu.StepLimit
				kb := uimenu.BuildAdminPromoBalanceWizardStep(lang, uimenu.StepLimit)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_limit_prompt")})
				return
			case uimenu.CallbackPromoAdminCancel:
				delete(h.adminStates, update.CallbackQuery.From.ID)
				h.clearAdminInputs(update.CallbackQuery.From.ID)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, Text: tm.GetText(lang, "promo_cancelled")})
				return
			}
		}
	case "sub":
		switch state.Step {
		case uimenu.StepCode:
			if data == uimenu.CallbackPromoAdminSubCodeRandom {
				state.Code = ""
				state.Step = uimenu.StepMonths
				kb := uimenu.BuildAdminPromoSubWizardStep(lang, uimenu.StepMonths)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_months_prompt")})
				return
			}
			if data == uimenu.CallbackPromoAdminSubCodeCustom {
				h.expectCode(update.CallbackQuery.From.ID)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, Text: tm.GetText(lang, "promo_code_prompt")})
				return
			}
		case uimenu.StepMonths:
			if strings.HasPrefix(data, uimenu.CallbackPromoAdminSubMonths) {
				val := strings.TrimPrefix(data, uimenu.CallbackPromoAdminSubMonths+":")
				m, _ := strconv.Atoi(val)
				state.Months = m
				state.Step = uimenu.StepLimit
				kb := uimenu.BuildAdminPromoSubWizardStep(lang, uimenu.StepLimit)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_limit_prompt")})
				return
			}
		case uimenu.StepLimit:
			if strings.HasPrefix(data, uimenu.CallbackPromoAdminSubLimit) {
				val := strings.TrimPrefix(data, uimenu.CallbackPromoAdminSubLimit+":")
				if val == "manual" {
					h.expectLimit(update.CallbackQuery.From.ID)
					_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, Text: tm.GetText(lang, "promo.limit.manual_prompt")})
				} else {
					l, _ := strconv.Atoi(val)
					state.Limit = l
					state.Step = uimenu.StepConfirm
					kb := uimenu.BuildAdminPromoSubWizardStep(lang, uimenu.StepConfirm)
					txt := fmt.Sprintf(tm.GetText(lang, "promo_confirm_text"), state.Months, state.Limit)
					_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: txt})
				}
				return
			}
		case uimenu.StepConfirm:
			switch data {
			case uimenu.CallbackPromoAdminSubConfirm:
				code := state.Code
				createdCode, err := h.promotionService.CreateSubscription(ctx, code, state.Months, state.Limit, update.CallbackQuery.From.ID)
				if err != nil {
					slog.Error("create sub promo", "err", err)
					return
				}
				fullCode := createdCode
				if !strings.HasPrefix(fullCode, "SUB_") {
					fullCode = "SUB_" + fullCode
				}
				text := fmt.Sprintf(tm.GetText(lang, "promo_sub_created"), fullCode, state.Months, state.Limit)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, Text: text})
				delete(h.adminStates, update.CallbackQuery.From.ID)
				h.clearAdminInputs(update.CallbackQuery.From.ID)
				return
			case uimenu.CallbackPromoAdminBack:
				state.Step = uimenu.StepLimit
				kb := uimenu.BuildAdminPromoSubWizardStep(lang, uimenu.StepLimit)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_limit_prompt")})
				return
			case uimenu.CallbackPromoAdminCancel:
				delete(h.adminStates, update.CallbackQuery.From.ID)
				h.clearAdminInputs(update.CallbackQuery.From.ID)
				_, _ = SafeEditMessageText(ctx, b, update.CallbackQuery.Message.Message, &bot.EditMessageTextParams{ChatID: chatID, MessageID: msgID, Text: tm.GetText(lang, "promo_cancelled")})
				return
			}
		}
	}
}

func (h *Handler) AdminPromoAmountMessageHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !contextkey.IsAdminFromContext(ctx) {
		return
	}
	if !h.consumeAmount(update.Message.Chat.ID) {
		return
	}
	lang := update.Message.From.LanguageCode
	tm := translation.GetInstance()
	state := h.adminStates[update.Message.From.ID]
	if state == nil {
		return
	}
	val, err := strconv.Atoi(strings.TrimSpace(update.Message.Text))
	if err != nil || val <= 0 {
		h.expectAmount(update.Message.From.ID)
		if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: tm.GetText(lang, "promo.balance.amount.invalid")}); serr != nil {
			slog.Error("send invalid amount", "err", serr)
		}
		return
	}
	state.Amount = val
	state.Step = uimenu.StepLimit
	kb := uimenu.BuildAdminPromoBalanceWizardStep(lang, uimenu.StepLimit)
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_limit_prompt")})
}

func (h *Handler) AdminPromoCodeMessageHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !contextkey.IsAdminFromContext(ctx) {
		return
	}
	if !h.consumeCode(update.Message.Chat.ID) {
		return
	}
	lang := update.Message.From.LanguageCode
	tm := translation.GetInstance()
	state := h.adminStates[update.Message.From.ID]
	if state == nil {
		return
	}
	code := strings.TrimSpace(update.Message.Text)
	if !promoCodeRe.MatchString(code) {
		h.expectCode(update.Message.From.ID)
		if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: tm.GetText(lang, "promo.code.invalid")}); serr != nil {
			slog.Error("send invalid code", "err", serr)
		}
		return
	}
	state.Code = code
	state.Step = uimenu.StepMonths
	kb := uimenu.BuildAdminPromoSubWizardStep(lang, uimenu.StepMonths)
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: tm.GetText(lang, "promo_months_prompt")})
}

func (h *Handler) AdminPromoLimitMessageHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !contextkey.IsAdminFromContext(ctx) {
		return
	}
	if !h.consumeLimit(update.Message.Chat.ID) {
		return
	}
	lang := update.Message.From.LanguageCode
	tm := translation.GetInstance()
	state := h.adminStates[update.Message.From.ID]
	if state == nil {
		return
	}
	val, err := strconv.Atoi(strings.TrimSpace(update.Message.Text))
	if err != nil || val < 0 {
		h.expectLimit(update.Message.From.ID)
		if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: tm.GetText(lang, "promo.limit.invalid")}); serr != nil {
			slog.Error("send invalid limit", "err", serr)
		}
		return
	}
	state.Limit = val
	state.Step = uimenu.StepConfirm
	var kb [][]models.InlineKeyboardButton
	var text string
	if state.Type == "balance" {
		kb = uimenu.BuildAdminPromoBalanceWizardStep(lang, uimenu.StepConfirm)
		text = fmt.Sprintf(tm.GetText(lang, "promo_confirm_text"), state.Amount, state.Limit)
	} else {
		kb = uimenu.BuildAdminPromoSubWizardStep(lang, uimenu.StepConfirm)
		text = fmt.Sprintf(tm.GetText(lang, "promo_confirm_text"), state.Months, state.Limit)
	}
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}, Text: text})
}
