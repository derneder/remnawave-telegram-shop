package promo

import "github.com/go-telegram/bot/models"

// BalanceAmountKeyboard returns buttons with preset amounts and manual input.
func BalanceAmountKeyboard() [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{
			{Text: "100", CallbackData: "promo_balance_amount:100"},
			{Text: "300", CallbackData: "promo_balance_amount:300"},
		},
		{
			{Text: "500", CallbackData: "promo_balance_amount:500"},
			{Text: "1000", CallbackData: "promo_balance_amount:1000"},
		},
		{
			{Text: "Ввести вручную", CallbackData: "promo_balance_amount:manual"},
		},
	}
}

// BalanceLimitKeyboard returns buttons with preset limits and manual input.
func BalanceLimitKeyboard() [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{
			{Text: "1", CallbackData: "promo_balance_limit:1"},
			{Text: "5", CallbackData: "promo_balance_limit:5"},
			{Text: "10", CallbackData: "promo_balance_limit:10"},
		},
		{
			{Text: "∞", CallbackData: "promo_balance_limit:0"},
			{Text: "Ввести вручную", CallbackData: "promo_balance_limit:manual"},
		},
	}
}
