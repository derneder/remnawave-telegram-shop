package handler

import (
	"context"
	"errors"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// inlineKeyboardEqual compares two inline keyboards for equality.
func inlineKeyboardEqual(a, b *models.InlineKeyboardMarkup) bool {
	if a == nil && b == nil {
		return true
	}
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a.InlineKeyboard) != len(b.InlineKeyboard) {
		return false
	}
	for i := range a.InlineKeyboard {
		if len(a.InlineKeyboard[i]) != len(b.InlineKeyboard[i]) {
			return false
		}
		for j := range a.InlineKeyboard[i] {
			if a.InlineKeyboard[i][j] != b.InlineKeyboard[i][j] {
				return false
			}
		}
	}
	return true
}

// SafeEditMessageText edits a message if new text or keyboard differ from the current ones.
// It also ignores the "message is not modified" telegram error.
func SafeEditMessageText(ctx context.Context, b *bot.Bot, current *models.Message, params *bot.EditMessageTextParams) (*models.Message, error) {
	var newMarkup *models.InlineKeyboardMarkup
	switch m := params.ReplyMarkup.(type) {
	case models.InlineKeyboardMarkup:
		newMarkup = &m
	case *models.InlineKeyboardMarkup:
		newMarkup = m
	}

	if current != nil {
		if current.Text == params.Text && inlineKeyboardEqual(current.ReplyMarkup, newMarkup) {
			return current, nil
		}
	}

	msg, err := b.EditMessageText(ctx, params)
	if err != nil {
		if errors.Is(err, bot.ErrorBadRequest) && strings.Contains(err.Error(), "message is not modified") {
			return current, nil
		}
		return nil, err
	}
	return msg, nil
}
