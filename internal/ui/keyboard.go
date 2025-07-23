package ui

import (
	"github.com/go-telegram/bot/models"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

func ConnectKeyboard(lang, backKey, backCallback string) [][]models.InlineKeyboardButton {
	tm := translation.GetInstance()
	var kb [][]models.InlineKeyboardButton
	if config.GetMiniAppURL() != "" {
		kb = append(kb, []models.InlineKeyboardButton{
			{Text: tm.GetText(lang, "connect_button"), WebApp: &models.WebAppInfo{URL: config.GetMiniAppURL()}},
		})
	} else {
		kb = append(kb, []models.InlineKeyboardButton{
			{Text: tm.GetText(lang, "connect_button"), CallbackData: "connect"},
		})
	}
	kb = append(kb, []models.InlineKeyboardButton{{Text: tm.GetText(lang, backKey), CallbackData: backCallback}})
	return kb
}
