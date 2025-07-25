package payment

import (
	"context"
	"fmt"
	"log/slog"
	"remnawave-tg-shop-bot/internal/adapter/remnawave"
	tg "remnawave-tg-shop-bot/internal/adapter/telegram/messenger"
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	domainpurchase "remnawave-tg-shop-bot/internal/domain/purchase"
	"remnawave-tg-shop-bot/internal/pkg/cache"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/contextkey"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	"remnawave-tg-shop-bot/internal/pkg/utils"
	"remnawave-tg-shop-bot/internal/repository/pg"
	custrepo "remnawave-tg-shop-bot/internal/service/customer"
	referralrepo "remnawave-tg-shop-bot/internal/service/referral"
	"remnawave-tg-shop-bot/internal/ui"
	"strings"
	"time"

	remapi "github.com/Jolymmiles/remnawave-api-go/api"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type PaymentService struct {
	repo                     PurchaseRepository
	remnawaveClient          *remnawave.Client
	customerRepository       custrepo.Repository
	messenger                tg.Messenger
	translation              *translation.Manager
	providers                map[domainpurchase.InvoiceType]Provider
	referralRepository       referralrepo.Repository
	promocodeRepository      *pg.PromocodeRepository
	promocodeUsageRepository *pg.PromocodeUsageRepository
	cache                    *cache.Cache
}

// EnabledProviders returns slice of active payment providers.
func (s PaymentService) EnabledProviders() []Provider {
	var res []Provider
	for _, p := range s.providers {
		if p.Enabled() {
			res = append(res, p)
		}
	}
	return res
}

func NewPaymentService(
	translation *translation.Manager,
	repo PurchaseRepository,
	remnawaveClient *remnawave.Client,
	customerRepository custrepo.Repository,
	messenger tg.Messenger,
	providers []Provider,
	referralRepository referralrepo.Repository,
	promocodeRepository *pg.PromocodeRepository,
	promocodeUsageRepository *pg.PromocodeUsageRepository,
	cache *cache.Cache,
) *PaymentService {
	provMap := make(map[domainpurchase.InvoiceType]Provider)
	for _, p := range providers {
		if p != nil {
			provMap[p.Type()] = p
		}
	}
	return &PaymentService{
		repo:                     repo,
		remnawaveClient:          remnawaveClient,
		customerRepository:       customerRepository,
		messenger:                messenger,
		translation:              translation,
		providers:                provMap,
		referralRepository:       referralRepository,
		promocodeRepository:      promocodeRepository,
		promocodeUsageRepository: promocodeUsageRepository,
		cache:                    cache,
	}
}

func (s PaymentService) ProcessPurchaseById(ctx context.Context, purchaseId int64) error {
	purchase, err := s.repo.FindById(ctx, purchaseId)
	if err != nil {
		return err
	}
	if purchase == nil {
		return fmt.Errorf("purchase with crypto invoice id %s not found", utils.MaskHalfInt64(purchaseId))
	}

	customer, err := s.customerRepository.FindById(ctx, purchase.CustomerID)
	if err != nil {
		return err
	}
	if customer == nil {
		return fmt.Errorf("customer %s not found", utils.MaskHalfInt64(purchase.CustomerID))
	}

	if messageId, b := s.cache.Get(purchase.ID); b {
		_, err = s.messenger.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    customer.TelegramID,
			MessageID: messageId,
		})
		if err != nil {
			slog.Error("Error deleting message", "err", err)
		}
		s.cache.Delete(purchase.ID)
	}

	err = s.repo.MarkAsPaid(ctx, purchase.ID)
	if err != nil {
		return err
	}

	newBalance := customer.Balance + purchase.Amount
	if err := s.customerRepository.UpdateFields(ctx, customer.ID, map[string]interface{}{"balance": newBalance}); err != nil {
		return err
	}
	customer.Balance = newBalance

	_, err = s.messenger.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: customer.TelegramID,
		Text:   fmt.Sprintf(s.translation.GetText(customer.Language, "balance_topped_up"), int(purchase.Amount)),
	})
	if err != nil {
		return err
	}

	if referral, err := s.referralRepository.FindByReferee(ctx, customer.TelegramID); err == nil && referral != nil && !referral.BonusGranted {
		referrer, err := s.customerRepository.FindByTelegramId(ctx, referral.ReferrerID)
		if err == nil && referrer != nil {
			bonus := float64(config.GetReferralBonus())
			newBal := referrer.Balance + bonus
			if err := s.customerRepository.UpdateFields(ctx, referrer.ID, map[string]interface{}{"balance": newBal}); err == nil {
				_ = s.referralRepository.MarkBonusGranted(ctx, referral.ID)
				if _, err := s.messenger.SendMessage(ctx, &bot.SendMessageParams{ChatID: referrer.TelegramID, Text: s.translation.GetText(referrer.Language, "referral_bonus_granted")}); err != nil {
					slog.Error("send referral bonus", "err", err)
				}
			}
		}
	}

	slog.Info("purchase processed", "purchase_id", utils.MaskHalfInt64(purchase.ID), "type", purchase.InvoiceType, "customer_id", utils.MaskHalfInt64(customer.ID))

	return nil
}

func (s PaymentService) PurchaseFromBalance(ctx context.Context, customer *domaincustomer.Customer, months int) error {
	price := config.Price(months)
	if customer.Balance < float64(price) {
		if _, err := s.messenger.SendMessage(ctx, &bot.SendMessageParams{ChatID: customer.TelegramID, Text: s.translation.GetText(customer.Language, "insufficient_balance")}); err != nil {
			slog.Error("send insufficient balance", "err", err)
		}
		return nil
	}

	user, err := s.remnawaveClient.CreateOrUpdateUser(ctx, customer.TelegramID, config.TrafficLimit(), months*30)
	if err != nil {
		return err
	}

	newBalance := customer.Balance - float64(price)
	updates := map[string]interface{}{
		"subscription_link": user.SubscriptionUrl,
		"expire_at":         user.ExpireAt,
		"balance":           newBalance,
	}

	customer.SubscriptionLink = &user.SubscriptionUrl
	customer.ExpireAt = &user.ExpireAt
	customer.Balance = newBalance

	if err := s.customerRepository.UpdateFields(ctx, customer.ID, updates); err != nil {
		return err
	}

	_, err = s.messenger.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      customer.TelegramID,
		ParseMode:   models.ParseModeHTML,
		Text:        s.translation.GetText(customer.Language, "subscription_activated"),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: ui.ConnectKeyboard(customer.Language, "back_button", "start")},
	})
	return err
}

func (s PaymentService) CreatePurchase(ctx context.Context, amount int, months int, customer *domaincustomer.Customer, invoiceType domainpurchase.InvoiceType) (url string, purchaseId int64, err error) {
	if customer == nil {
		return "", 0, fmt.Errorf("customer is nil")
	}
	switch invoiceType {
	case domainpurchase.InvoiceTypeCrypto:
		if p, ok := s.providers[domainpurchase.InvoiceTypeCrypto]; ok {
			return p.CreateInvoice(ctx, amount, months, customer)
		}
		return "", 0, fmt.Errorf("unknown invoice type: %s", invoiceType)
	case domainpurchase.InvoiceTypeTelegram:
		return s.createTelegramInvoice(ctx, amount, months, customer)
	case domainpurchase.InvoiceTypeTribute:
		if p, ok := s.providers[domainpurchase.InvoiceTypeTribute]; ok {
			return p.CreateInvoice(ctx, amount, months, customer)
		}
		return "", 0, fmt.Errorf("unknown invoice type: %s", invoiceType)
	default:
		return "", 0, fmt.Errorf("unknown invoice type: %s", invoiceType)
	}
}

func (s PaymentService) createTelegramInvoice(ctx context.Context, amount int, months int, customer *domaincustomer.Customer) (url string, purchaseId int64, err error) {
	purchaseId, err = s.repo.Create(ctx, &domainpurchase.Purchase{
		InvoiceType: domainpurchase.InvoiceTypeTelegram,
		Status:      domainpurchase.StatusNew,
		Amount:      float64(amount),
		Currency:    "STARS",
		CustomerID:  customer.ID,
		Month:       months,
	})
	if err != nil {
		slog.Error("Error creating purchase", "err", err)
		return "", 0, nil
	}

	invoiceUrl, err := s.messenger.CreateInvoiceLink(ctx, &bot.CreateInvoiceLinkParams{
		Title:    s.translation.GetText(customer.Language, "invoice_title"),
		Currency: "XTR",
		Prices: []models.LabeledPrice{
			{
				Label:  s.translation.GetText(customer.Language, "invoice_label"),
				Amount: amount,
			},
		},
		Description: s.translation.GetText(customer.Language, "invoice_description"),
		Payload:     fmt.Sprintf("%d&%s", purchaseId, ctx.Value(contextkey.Username)),
	})

	if err != nil {
		slog.Error("Error creating stars invoice", "err", err)
		return "", 0, err
	}

	updates := map[string]interface{}{
		"status": domainpurchase.StatusPending,
	}

	err = s.repo.UpdateFields(ctx, purchaseId, updates)
	if err != nil {
		slog.Error("Error updating purchase", "err", err)
		return "", 0, err
	}

	return invoiceUrl, purchaseId, nil
}

func (s PaymentService) ActivateTrial(ctx context.Context, telegramId int64) (string, error) {
	if config.TrialDays() == 0 {
		return "", nil
	}
	customer, err := s.customerRepository.FindByTelegramId(ctx, telegramId)
	if err != nil {
		slog.Error("Error finding customer", "err", err)
		return "", err
	}
	if customer == nil {
		return "", fmt.Errorf("customer %d not found", telegramId)
	}
	user, err := s.remnawaveClient.CreateOrUpdateUser(ctx, telegramId, config.TrialTrafficLimit(), config.TrialDays())
	if err != nil {
		slog.Error("Error creating user", "err", err)
		return "", err
	}

	customerFilesToUpdate := map[string]interface{}{
		"subscription_link": user.GetSubscriptionUrl(),
		"expire_at":         user.GetExpireAt(),
	}

	err = s.customerRepository.UpdateFields(ctx, customer.ID, customerFilesToUpdate)
	if err != nil {
		return "", err
	}

	return user.GetSubscriptionUrl(), nil

}

func (s PaymentService) CancelPayment(purchaseId int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	purchase, err := s.repo.FindById(ctx, purchaseId)
	if err != nil {
		return err
	}
	if purchase == nil {
		return fmt.Errorf("purchase with crypto invoice id %s not found", utils.MaskHalfInt64(purchaseId))
	}

	purchaseFieldsToUpdate := map[string]interface{}{
		"status": domainpurchase.StatusCancel,
	}

	err = s.repo.UpdateFields(ctx, purchaseId, purchaseFieldsToUpdate)
	if err != nil {
		return err
	}

	return nil
}

func (s PaymentService) GetUser(ctx context.Context, telegramId int64) (*remapi.UserDto, error) {
	return s.remnawaveClient.GetUserByTelegramID(ctx, telegramId)
}

func (s PaymentService) GetUserDailyUsage(ctx context.Context, uuid string, start, end time.Time) (float64, error) {
	return s.remnawaveClient.GetUserDailyUsage(ctx, uuid, start, end)
}
func (s PaymentService) CreatePromocode(ctx context.Context, customer *domaincustomer.Customer, months, uses int) (string, error) {
	cost := config.Price(months) * uses
	if !config.IsAdmin(customer.TelegramID) {
		if customer.Balance < float64(cost) {
			return "", fmt.Errorf("insufficient balance")
		}
		newBalance := customer.Balance - float64(cost)
		if err := s.customerRepository.UpdateFields(ctx, customer.ID, map[string]interface{}{"balance": newBalance}); err != nil {
			return "", err
		}
		customer.Balance = newBalance
	}

	tmpCode := uuid.New().String()
	tmpCode = strings.ReplaceAll(tmpCode, "-", "")
	var code string
	for i, r := range tmpCode {
		if i%5 == 0 {
			code = fmt.Sprintf("%s-", code)
		}
		code = fmt.Sprintf("%s%c", code, r)
	}

	code = code[1:24]

	_, err := s.promocodeRepository.Create(ctx, &pg.Promocode{
		Code:      code,
		Months:    months,
		Type:      1,
		Days:      months * 30,
		UsesLeft:  uses,
		CreatedBy: customer.TelegramID,
		Active:    true,
	})
	if err != nil {
		return "", err
	}
	return code, nil
}

func (s PaymentService) ApplyPromocode(ctx context.Context, customer *domaincustomer.Customer, code string) error {
	promo, err := s.promocodeRepository.GetByCode(ctx, code)
	if err != nil {
		return err
	}
	if promo == nil || (!promo.Active) || promo.Deleted || (promo.UsesLeft <= 0 && promo.UsesLeft != 0) {
		return fmt.Errorf("invalid promocode")
	}
	if promo.Type == 2 {
		newBal := customer.Balance + float64(promo.Amount)
		if err := s.customerRepository.UpdateFields(ctx, customer.ID, map[string]interface{}{"balance": newBal}); err != nil {
			return err
		}
		customer.Balance = newBal
	} else {
		days := promo.Days
		if days == 0 {
			days = promo.Months * 30
		}
		user, err := s.remnawaveClient.CreateOrUpdateUser(ctx, customer.TelegramID, config.TrafficLimit(), days)
		if err != nil {
			return err
		}
		updates := map[string]interface{}{
			"subscription_link": user.SubscriptionUrl,
			"expire_at":         user.ExpireAt,
		}
		if err := s.customerRepository.UpdateFields(ctx, customer.ID, updates); err != nil {
			return err
		}
		customer.SubscriptionLink = &user.SubscriptionUrl
		customer.ExpireAt = &user.ExpireAt
	}

	if promo.UsesLeft > 0 {
		if err := s.promocodeRepository.DecrementUses(ctx, promo.ID); err != nil {
			return err
		}
	}
	_ = s.promocodeUsageRepository.Create(ctx, promo.ID, customer.TelegramID)
	return nil
}

func (s PaymentService) SetPromocodeStatus(ctx context.Context, id int64, active bool) error {
	return s.promocodeRepository.UpdateStatus(ctx, id, active)
}
