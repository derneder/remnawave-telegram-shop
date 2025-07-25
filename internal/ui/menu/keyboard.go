package menu

import (
	"time"

	"github.com/go-telegram/bot/models"
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	"remnawave-tg-shop-bot/internal/ui"
)

// Callback data constants for admin menu and promo wizard.
const (
	CallbackAdminMenu              = "admin_menu"
	CallbackAdminPromoBalanceStart = "admin_promo_balance_start"
	CallbackAdminPromoSubStart     = "admin_promo_sub_start"
	CallbackPromoBalanceAmount     = "promo_balance_amount"
	CallbackPromoBalanceLimit      = "promo_balance_limit"
	CallbackPromoBalanceConfirm    = "promo_balance_confirm"
	CallbackPromoSubCodeRandom     = "promo_sub_code_random"
	CallbackPromoSubCodeCustom     = "promo_sub_code_custom"
	CallbackPromoSubDays           = "promo_sub_days"
	CallbackPromoSubLimit          = "promo_sub_limit"
	CallbackPromoSubConfirm        = "promo_sub_confirm"
	CallbackAdminBack              = "admin_back"
	CallbackAdminCancel            = "admin_cancel"
)

// StepState represents wizard step identifier.
// For balance promo: Amount -> Limit -> Confirm.
// For subscription promo: Code -> Days -> Limit -> Confirm.
type StepState int

const (
	StepAmount StepState = iota
	StepLimit
	StepConfirm
	StepCode
	StepDays
)

// BuildMainKeyboard creates main menu keyboard.
func BuildMainKeyboard(lang string, c *domaincustomer.Customer, isAdmin bool) [][]models.InlineKeyboardButton {
	tm := translation.GetInstance()
	var kb [][]models.InlineKeyboardButton
	if c.SubscriptionLink == nil && config.TrialDays() > 0 {
		kb = append(kb, []models.InlineKeyboardButton{{Text: tm.GetText(lang, "trial_button"), CallbackData: "trial"}})
	}
	kb = append(kb, []models.InlineKeyboardButton{{Text: tm.GetText(lang, "refresh_button"), CallbackData: "start"}})
	kb = append(kb, []models.InlineKeyboardButton{{Text: tm.GetText(lang, "balance_menu_button"), CallbackData: "balance"}})
	if c.SubscriptionLink != nil && c.ExpireAt.After(time.Now()) {
		kb = append(kb, []models.InlineKeyboardButton{ui.MakeConnectButton(config.GetMiniAppURL(), lang)})
	}
	kb = append(kb, []models.InlineKeyboardButton{{Text: tm.GetText(lang, "referral_button"), CallbackData: "referral"}})
	kb = append(kb, []models.InlineKeyboardButton{{Text: tm.GetText(lang, "other_button"), CallbackData: "other"}})
	if isAdmin {
		kb = append(kb, []models.InlineKeyboardButton{{Text: tm.GetText(lang, "admin_panel_button"), CallbackData: CallbackAdminMenu}})
	}
	var row []models.InlineKeyboardButton
	if config.SupportURL() != "" {
		row = append(row, models.InlineKeyboardButton{Text: tm.GetText(lang, "support_button"), URL: config.SupportURL()})
	}
	if config.ChannelURL() != "" {
		row = append(row, models.InlineKeyboardButton{Text: tm.GetText(lang, "channel_button"), URL: config.ChannelURL()})
	}
	if len(row) > 0 {
		kb = append(kb, row)
	}
	return kb
}

// BuildRefPromoUserMenu returns referral & promo menu for regular users.
func BuildRefPromoUserMenu(lang string) [][]models.InlineKeyboardButton {
	tm := translation.GetInstance()
	return [][]models.InlineKeyboardButton{
		{{Text: tm.GetText(lang, "referral_system_button"), CallbackData: "referral_stats"}},
		{{Text: tm.GetText(lang, "enter_promocode_button"), CallbackData: "promo_enter"}},
		{{Text: tm.GetText(lang, "faq_button"), CallbackData: "faq"}},
		{{Text: tm.GetText(lang, "back_to_account_button"), CallbackData: "start"}},
	}
}

// BuildRefPromoAdminMenu returns referral & promo menu for admins.
func BuildRefPromoAdminMenu(lang string) [][]models.InlineKeyboardButton {
	kb := BuildRefPromoUserMenu(lang)
	tm := translation.GetInstance()
	return append([][]models.InlineKeyboardButton{{{Text: tm.GetText(lang, "admin_panel_button"), CallbackData: CallbackAdminMenu}}}, kb...)
}

// BuildAdminPromoMenu returns root admin promo menu keyboard.
func BuildAdminPromoMenu(lang string) [][]models.InlineKeyboardButton {
	tm := translation.GetInstance()
	return [][]models.InlineKeyboardButton{
		{{Text: tm.GetText(lang, "admin_promo_balance_button"), CallbackData: CallbackAdminPromoBalanceStart}},
		{{Text: tm.GetText(lang, "admin_promo_sub_button"), CallbackData: CallbackAdminPromoSubStart}},
		{{Text: tm.GetText(lang, "back_button"), CallbackData: "start"}},
	}
}

// BuildAdminPromoBalanceWizardStep builds keyboards for balance promo creation wizard.
func BuildAdminPromoBalanceWizardStep(lang string, step StepState) [][]models.InlineKeyboardButton {
	tm := translation.GetInstance()
	switch step {
	case StepAmount:
		return [][]models.InlineKeyboardButton{
			{{Text: "100", CallbackData: CallbackPromoBalanceAmount + ":100"}, {Text: "300", CallbackData: CallbackPromoBalanceAmount + ":300"}},
			{{Text: "500", CallbackData: CallbackPromoBalanceAmount + ":500"}, {Text: "1000", CallbackData: CallbackPromoBalanceAmount + ":1000"}},
			{{Text: tm.GetText(lang, "manual_input_button"), CallbackData: CallbackPromoBalanceAmount + ":manual"}},
			{{Text: tm.GetText(lang, "cancel_button"), CallbackData: CallbackAdminCancel}},
		}
	case StepLimit:
		return [][]models.InlineKeyboardButton{
			{{Text: "1", CallbackData: CallbackPromoBalanceLimit + ":1"}, {Text: "5", CallbackData: CallbackPromoBalanceLimit + ":5"}, {Text: "10", CallbackData: CallbackPromoBalanceLimit + ":10"}},
			{{Text: "∞", CallbackData: CallbackPromoBalanceLimit + ":0"}, {Text: tm.GetText(lang, "manual_input_button"), CallbackData: CallbackPromoBalanceLimit + ":manual"}},
			{{Text: tm.GetText(lang, "back_button"), CallbackData: CallbackAdminBack}},
		}
	case StepConfirm:
		return [][]models.InlineKeyboardButton{
			{{Text: tm.GetText(lang, "create_button"), CallbackData: CallbackPromoBalanceConfirm}},
			{{Text: tm.GetText(lang, "back_button"), CallbackData: CallbackAdminBack}},
			{{Text: tm.GetText(lang, "cancel_button"), CallbackData: CallbackAdminCancel}},
		}
	default:
		return nil
	}
}

// BuildAdminPromoSubWizardStep builds keyboards for subscription promo creation wizard.
func BuildAdminPromoSubWizardStep(lang string, step StepState) [][]models.InlineKeyboardButton {
	tm := translation.GetInstance()
	switch step {
	case StepCode:
		return [][]models.InlineKeyboardButton{
			{{Text: "🎲", CallbackData: CallbackPromoSubCodeRandom}, {Text: "✍️", CallbackData: CallbackPromoSubCodeCustom}},
			{{Text: tm.GetText(lang, "cancel_button"), CallbackData: CallbackAdminCancel}},
		}
	case StepDays:
		return [][]models.InlineKeyboardButton{
			{{Text: "30", CallbackData: CallbackPromoSubDays + ":30"}, {Text: "90", CallbackData: CallbackPromoSubDays + ":90"}},
			{{Text: "180", CallbackData: CallbackPromoSubDays + ":180"}, {Text: "365", CallbackData: CallbackPromoSubDays + ":365"}},
			{{Text: tm.GetText(lang, "back_button"), CallbackData: CallbackAdminBack}},
		}
	case StepLimit:
		return [][]models.InlineKeyboardButton{
			{{Text: "1", CallbackData: CallbackPromoSubLimit + ":1"}, {Text: "5", CallbackData: CallbackPromoSubLimit + ":5"}, {Text: "10", CallbackData: CallbackPromoSubLimit + ":10"}},
			{{Text: "∞", CallbackData: CallbackPromoSubLimit + ":0"}, {Text: tm.GetText(lang, "manual_input_button"), CallbackData: CallbackPromoSubLimit + ":manual"}},
			{{Text: tm.GetText(lang, "back_button"), CallbackData: CallbackAdminBack}},
		}
	case StepConfirm:
		return [][]models.InlineKeyboardButton{
			{{Text: tm.GetText(lang, "create_button"), CallbackData: CallbackPromoSubConfirm}},
			{{Text: tm.GetText(lang, "back_button"), CallbackData: CallbackAdminBack}},
			{{Text: tm.GetText(lang, "cancel_button"), CallbackData: CallbackAdminCancel}},
		}
	default:
		return nil
	}
}
