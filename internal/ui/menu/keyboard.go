package menu

import (
	"time"

	"github.com/go-telegram/bot/models"
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	"remnawave-tg-shop-bot/internal/ui"
)

// Callback data constants for promo/referral menus and admin promo wizard.
const (
	// user callbacks
	CallbackPromoUserActivate    = "promo_user_activate"
	CallbackRefUserStats         = "ref_user_stats"
	CallbackPromoUserPersonal    = "promo_user_personal"
	CallbackPromoMyList          = "promo_my_list"
	CallbackPromoMyFreeze        = "promo_my_freeze"
	CallbackPromoMyUnfreeze      = "promo_my_unfreeze"
	CallbackPromoMyDelete        = "promo_my_delete"
	CallbackPromoMyDeleteConfirm = "promo_my_delete_confirm"
	// admin callbacks
	CallbackPromoAdminMenu           = "promo_admin_menu"
	CallbackPromoAdminBalanceStart   = "promo_admin_balance_start"
	CallbackPromoAdminSubStart       = "promo_admin_sub_start"
	CallbackPromoAdminBalanceAmount  = "promo_admin_balance_amount"
	CallbackPromoAdminBalanceLimit   = "promo_admin_balance_limit"
	CallbackPromoAdminBalanceConfirm = "promo_admin_balance_confirm"
	CallbackPromoAdminSubCodeRandom  = "promo_admin_sub_code_random"
	CallbackPromoAdminSubCodeCustom  = "promo_admin_sub_code_custom"
	CallbackPromoAdminSubDays        = "promo_admin_sub_days"
	CallbackPromoAdminSubLimit       = "promo_admin_sub_limit"
	CallbackPromoAdminSubConfirm     = "promo_admin_sub_confirm"
	CallbackPromoAdminBack           = "promo_admin_back"
	CallbackPromoAdminCancel         = "promo_admin_cancel"
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

// BuildLKMenu creates personal account main menu.
func BuildLKMenu(lang string, c *domaincustomer.Customer, isAdmin bool) [][]models.InlineKeyboardButton {
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
		kb = append(kb, []models.InlineKeyboardButton{{Text: tm.GetText(lang, "admin_panel_button"), CallbackData: CallbackPromoAdminMenu}})
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

// BuildPromoRefMenu builds promo/referral menu for regular users.
// BuildPromoRefMain builds promo/referral menu for regular users.
// If isAdmin is true the admin promo menu button is also added.
func BuildPromoRefMain(lang string, isAdmin bool) [][]models.InlineKeyboardButton {
	tm := translation.GetInstance()
	kb := [][]models.InlineKeyboardButton{
		{{Text: tm.GetText(lang, "activate_promocode_button"), CallbackData: CallbackPromoUserActivate}},
		{{Text: tm.GetText(lang, "referral_system_button"), CallbackData: CallbackRefUserStats}},
		{{Text: tm.GetText(lang, "personal_promocodes_button"), CallbackData: CallbackPromoMyList}},
	}
	if isAdmin {
		kb = append(kb, []models.InlineKeyboardButton{{Text: tm.GetText(lang, "admin_panel_button"), CallbackData: CallbackPromoAdminMenu}})
	}
	kb = append(kb, []models.InlineKeyboardButton{{Text: tm.GetText(lang, "back_to_account_button"), CallbackData: "start"}})
	return kb
}

// BuildPromoRefMenu is kept for backward compatibility.
func BuildPromoRefMenu(lang string) [][]models.InlineKeyboardButton {
	return BuildPromoRefMain(lang, false)
}

// BuildAdminPromoMenu returns root admin promo menu keyboard.
func BuildAdminPromoMenu(lang string) [][]models.InlineKeyboardButton {
	tm := translation.GetInstance()
	return [][]models.InlineKeyboardButton{
		{{Text: tm.GetText(lang, "admin_promo_balance_button"), CallbackData: CallbackPromoAdminBalanceStart}},
		{{Text: tm.GetText(lang, "admin_promo_sub_button"), CallbackData: CallbackPromoAdminSubStart}},
		{{Text: tm.GetText(lang, "back_button"), CallbackData: "start"}},
	}
}

// BuildAdminPromoBalanceWizardStep builds keyboards for balance promo creation wizard.
func BuildAdminPromoBalanceWizardStep(lang string, step StepState) [][]models.InlineKeyboardButton {
	tm := translation.GetInstance()
	switch step {
	case StepAmount:
		return [][]models.InlineKeyboardButton{
			{{Text: "100", CallbackData: CallbackPromoAdminBalanceAmount + ":100"}, {Text: "300", CallbackData: CallbackPromoAdminBalanceAmount + ":300"}},
			{{Text: "500", CallbackData: CallbackPromoAdminBalanceAmount + ":500"}, {Text: "1000", CallbackData: CallbackPromoAdminBalanceAmount + ":1000"}},
			{{Text: tm.GetText(lang, "manual_input_button"), CallbackData: CallbackPromoAdminBalanceAmount + ":manual"}},
			{{Text: tm.GetText(lang, "cancel_button"), CallbackData: CallbackPromoAdminCancel}},
		}
	case StepLimit:
		return [][]models.InlineKeyboardButton{
			{{Text: "1", CallbackData: CallbackPromoAdminBalanceLimit + ":1"}, {Text: "5", CallbackData: CallbackPromoAdminBalanceLimit + ":5"}, {Text: "10", CallbackData: CallbackPromoAdminBalanceLimit + ":10"}},
			{{Text: "∞", CallbackData: CallbackPromoAdminBalanceLimit + ":0"}, {Text: tm.GetText(lang, "manual_input_button"), CallbackData: CallbackPromoAdminBalanceLimit + ":manual"}},
			{{Text: tm.GetText(lang, "back_button"), CallbackData: CallbackPromoAdminBack}},
		}
	case StepConfirm:
		return [][]models.InlineKeyboardButton{
			{{Text: tm.GetText(lang, "create_button"), CallbackData: CallbackPromoAdminBalanceConfirm}},
			{{Text: tm.GetText(lang, "back_button"), CallbackData: CallbackPromoAdminBack}},
			{{Text: tm.GetText(lang, "cancel_button"), CallbackData: CallbackPromoAdminCancel}},
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
			{{Text: "🎲", CallbackData: CallbackPromoAdminSubCodeRandom}, {Text: "✍️", CallbackData: CallbackPromoAdminSubCodeCustom}},
			{{Text: tm.GetText(lang, "cancel_button"), CallbackData: CallbackPromoAdminCancel}},
		}
	case StepDays:
		return [][]models.InlineKeyboardButton{
			{{Text: "30", CallbackData: CallbackPromoAdminSubDays + ":30"}, {Text: "90", CallbackData: CallbackPromoAdminSubDays + ":90"}},
			{{Text: "180", CallbackData: CallbackPromoAdminSubDays + ":180"}, {Text: "365", CallbackData: CallbackPromoAdminSubDays + ":365"}},
			{{Text: tm.GetText(lang, "back_button"), CallbackData: CallbackPromoAdminBack}},
		}
	case StepLimit:
		return [][]models.InlineKeyboardButton{
			{{Text: "1", CallbackData: CallbackPromoAdminSubLimit + ":1"}, {Text: "5", CallbackData: CallbackPromoAdminSubLimit + ":5"}, {Text: "10", CallbackData: CallbackPromoAdminSubLimit + ":10"}},
			{{Text: "∞", CallbackData: CallbackPromoAdminSubLimit + ":0"}, {Text: tm.GetText(lang, "manual_input_button"), CallbackData: CallbackPromoAdminSubLimit + ":manual"}},
			{{Text: tm.GetText(lang, "back_button"), CallbackData: CallbackPromoAdminBack}},
		}
	case StepConfirm:
		return [][]models.InlineKeyboardButton{
			{{Text: tm.GetText(lang, "create_button"), CallbackData: CallbackPromoAdminSubConfirm}},
			{{Text: tm.GetText(lang, "back_button"), CallbackData: CallbackPromoAdminBack}},
			{{Text: tm.GetText(lang, "cancel_button"), CallbackData: CallbackPromoAdminCancel}},
		}
	default:
		return nil
	}
}
