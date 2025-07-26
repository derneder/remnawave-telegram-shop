package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnawave-tg-shop-bot/internal/pkg/translation"
	menu "remnawave-tg-shop-bot/internal/ui/menu"
)

type personalStep int

const (
	personalStepMonths personalStep = iota
	personalStepUses
)

type personalState struct {
	Step   personalStep
	Months int
}

func (h *Handler) PersonalCodesCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	lang := update.CallbackQuery.From.LanguageCode
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	kb := menu.BuildPersonalCodesMenu(lang)
	tm := translation.GetInstance()
	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: msgID,
		ParseMode: models.ParseModeHTML,
		Text:      tm.GetText(lang, "personal_codes_button"),
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: kb,
		},
	})
	if err != nil {
		slog.Error("send personal menu", "err", err)
	}
}

func (h *Handler) PersonalCreateCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	lang := update.CallbackQuery.From.LanguageCode
	chatID, msgID, err := getCallbackIDs(update)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	h.promoMu.Lock()
	h.personalStates[update.CallbackQuery.From.ID] = &personalState{Step: personalStepMonths}
	h.promoMu.Unlock()
	var curMsg *models.Message
	if update.CallbackQuery.Message.Message != nil {
		curMsg = update.CallbackQuery.Message.Message
	}
	_, err = SafeEditMessageText(ctx, b, curMsg, &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: msgID,
		Text:      h.translation.GetText(lang, "personal_months_prompt"),
	})
	if err != nil {
		slog.Error("send personal months", "err", err)
	}
}

func (h *Handler) IsAwaitingPersonalMonths(id int64) bool {
	h.promoMu.RLock()
	defer h.promoMu.RUnlock()
	st, ok := h.personalStates[id]
	return ok && st.Step == personalStepMonths
}

func (h *Handler) IsAwaitingPersonalUses(id int64) bool {
	h.promoMu.RLock()
	defer h.promoMu.RUnlock()
	st, ok := h.personalStates[id]
	return ok && st.Step == personalStepUses
}

func (h *Handler) personalState(id int64) *personalState {
	h.promoMu.RLock()
	defer h.promoMu.RUnlock()
	return h.personalStates[id]
}

func (h *Handler) setPersonalState(id int64, st *personalState) {
	h.promoMu.Lock()
	h.personalStates[id] = st
	h.promoMu.Unlock()
}

func (h *Handler) clearPersonalState(id int64) {
	h.promoMu.Lock()
	delete(h.personalStates, id)
	h.promoMu.Unlock()
}

func (h *Handler) PersonalMessageHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	st := h.personalState(update.Message.Chat.ID)
	if st == nil {
		return
	}
	lang := update.Message.From.LanguageCode
	txt := strings.TrimSpace(update.Message.Text)
	switch st.Step {
	case personalStepMonths:
		m, err := strconv.Atoi(txt)
		if err != nil || m <= 0 {
			h.setPersonalState(update.Message.Chat.ID, st)
			if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: h.translation.GetText(lang, "promo.limit.invalid")}); serr != nil {
				slog.Error("send invalid months", "err", serr)
			}
			return
		}
		st.Months = m
		st.Step = personalStepUses
		h.setPersonalState(update.Message.Chat.ID, st)
		if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: h.translation.GetText(lang, "personal_uses_prompt")}); serr != nil {
			slog.Error("send uses prompt", "err", serr)
		}
	case personalStepUses:
		u, err := strconv.Atoi(txt)
		if err != nil || u < 0 {
			h.setPersonalState(update.Message.Chat.ID, st)
			if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: h.translation.GetText(lang, "promo.limit.invalid")}); serr != nil {
				slog.Error("send invalid uses", "err", serr)
			}
			return
		}
		customer, err := h.findOrCreateCustomer(ctx, update.Message.Chat.ID, lang)
		if err != nil {
			slog.Error("find customer", "err", err)
			return
		}
		code := ""
		if h.paymentService != nil {
			code, err = h.paymentService.CreatePromocode(ctx, customer, st.Months, u)
			if err != nil {
				slog.Error("create promocode", "err", err)
				return
			}
		}
		if code == "" {
			code = "-"
		}
		if _, serr := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: update.Message.Chat.ID, Text: fmt.Sprintf(h.translation.GetText(lang, "personal_created_text"), code)}); serr != nil {
			slog.Error("send code", "err", serr)
		}
		h.clearPersonalState(update.Message.Chat.ID)
		// optionally show list after creation
	}
}
