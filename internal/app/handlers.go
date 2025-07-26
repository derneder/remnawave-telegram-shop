package app

import (
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"remnawave-tg-shop-bot/internal/adapter/telegram/handler"
	uimenu "remnawave-tg-shop-bot/internal/ui/menu"
)

func (a *App) InitHandlers(h *handler.Handler) {
	b := a.Bot
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, h.StartCommandHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/menu", bot.MatchTypeExact, h.MenuCommandHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, h.HelpCommandHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/promo", bot.MatchTypeExact, h.PromoCommandHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/promocode", bot.MatchTypeExact, h.PromocodeCommandHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/connect", bot.MatchTypeExact, h.ConnectCommandHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/sync", bot.MatchTypeExact, h.SyncUsersCommandHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackStart, bot.MatchTypePrefix, h.StartCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackConnect, bot.MatchTypePrefix, h.ConnectCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackBuy, bot.MatchTypePrefix, h.BuyCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackSell, bot.MatchTypePrefix, h.SellCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackPayment, bot.MatchTypePrefix, h.PaymentCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackBalance, bot.MatchTypePrefix, h.BalanceCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackTopup, bot.MatchTypePrefix, h.TopupCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackTopupMethod, bot.MatchTypePrefix, h.TopupMethodCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackPayFromBal, bot.MatchTypePrefix, h.PayFromBalanceCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackTrial, bot.MatchTypePrefix, h.TrialCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackActivateTrial, bot.MatchTypePrefix, h.ActivateTrialCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackReferral, bot.MatchTypePrefix, h.ReferralCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackReferralStats, bot.MatchTypePrefix, h.ReferralStatsCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, uimenu.CallbackRefUserStats, bot.MatchTypePrefix, h.ReferralStatsCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, uimenu.CallbackPromoAdminMenu, bot.MatchTypePrefix, h.AdminPromoCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "promo_admin_balance_", bot.MatchTypePrefix, h.AdminPromoCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "promo_admin_sub_", bot.MatchTypePrefix, h.AdminPromoCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackPromoEnter, bot.MatchTypePrefix, h.PromoEnterCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, uimenu.CallbackPromoUserActivate, bot.MatchTypePrefix, h.PromoEnterCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, uimenu.CallbackPromoMyList, bot.MatchTypePrefix, h.PromoMyListCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackOther, bot.MatchTypePrefix, h.OtherCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackFAQ, bot.MatchTypePrefix, h.FAQCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackTrafficLimit, bot.MatchTypePrefix, h.TrafficLimitCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackKeys, bot.MatchTypePrefix, h.KeysCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackQR, bot.MatchTypePrefix, h.QRCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackShortLink, bot.MatchTypePrefix, h.ShortLinkCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackShortList, bot.MatchTypePrefix, h.ShortListCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackLocations, bot.MatchTypePrefix, h.LocationsCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackRegenKey, bot.MatchTypePrefix, h.RegenKeyCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackProxy, bot.MatchTypePrefix, h.ProxyCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackLanguage, bot.MatchTypePrefix, h.LanguageCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackSetLanguage, bot.MatchTypePrefix, h.SetLanguageCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "", bot.MatchTypePrefix, h.UnknownCallbackHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)

	b.RegisterHandlerMatchFunc(func(upd *models.Update) bool {
		if upd.Message == nil {
			return false
		}
		return h.IsAwaitingPromo(upd.Message.Chat.ID)
	}, h.PromoCodeMessageHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)

	b.RegisterHandlerMatchFunc(func(upd *models.Update) bool {
		if upd.Message == nil {
			return false
		}
		return h.IsAwaitingAmount(upd.Message.Chat.ID)
	}, h.AdminPromoAmountMessageHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)

	b.RegisterHandlerMatchFunc(func(upd *models.Update) bool {
		if upd.Message == nil {
			return false
		}
		return h.IsAwaitingCode(upd.Message.Chat.ID)
	}, h.AdminPromoCodeMessageHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)

	b.RegisterHandlerMatchFunc(func(upd *models.Update) bool {
		if upd.Message == nil {
			return false
		}
		return h.IsAwaitingLimit(upd.Message.Chat.ID)
	}, h.AdminPromoLimitMessageHandler, h.CreateCustomerIfNotExistMiddleware, handler.LogUpdateMiddleware)

}
