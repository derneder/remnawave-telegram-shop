package ui

import (
	"github.com/go-telegram/bot/models"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

func makeConnectButton(miniAppURL, lang string) models.InlineKeyboardButton {
	tm := translation.GetInstance()
	if miniAppURL == "" {
		return models.InlineKeyboardButton{Text: tm.GetText(lang, "connect_button"), CallbackData: "connect"}
	}
	return models.InlineKeyboardButton{Text: tm.GetText(lang, "connect_button"), URL: miniAppURL}
}

func ConnectKeyboard(lang, backKey, backCallback string) [][]models.InlineKeyboardButton {
	tm := translation.GetInstance()
	var kb [][]models.InlineKeyboardButton
	kb = append(kb, []models.InlineKeyboardButton{makeConnectButton(config.GetMiniAppURL(), lang)})
	kb = append(kb, []models.InlineKeyboardButton{{Text: tm.GetText(lang, backKey), CallbackData: backCallback}})
	return kb
}
