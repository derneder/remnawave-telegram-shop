package payment

import (
	"context"
	"fmt"
	"log/slog"

	"remnawave-tg-shop-bot/internal/adapter/payment/cryptopay"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/contextkey"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	domainpurchase "remnawave-tg-shop-bot/internal/domain/purchase"
)

// CryptoPayProvider implements Provider for the CryptoPay service.
type CryptoPayProvider struct {
	repo   PurchaseRepository
	client *cryptopay.Client
}

func NewCryptoPayProvider(repo PurchaseRepository, client *cryptopay.Client) *CryptoPayProvider {
	return &CryptoPayProvider{repo: repo, client: client}
}

func (p CryptoPayProvider) Type() domainpurchase.InvoiceType { return domainpurchase.InvoiceTypeCrypto }

func (p CryptoPayProvider) Enabled() bool { return config.IsCryptoPayEnabled() }

func (p CryptoPayProvider) CreateInvoice(ctx context.Context, amount int, months int, customer *domaincustomer.Customer) (string, int64, error) {
	purchaseID, err := p.repo.Create(ctx, &domainpurchase.Purchase{
		InvoiceType: domainpurchase.InvoiceTypeCrypto,
		Status:      domainpurchase.StatusNew,
		Amount:      float64(amount),
		Currency:    "RUB",
		CustomerID:  customer.ID,
		Month:       months,
	})
	if err != nil {
		slog.Error("Error creating purchase", "err", err)
		return "", 0, err
	}

	invoice, err := p.client.CreateInvoice(&cryptopay.InvoiceRequest{
		CurrencyType:   "fiat",
		Fiat:           "RUB",
		Amount:         fmt.Sprintf("%d", amount),
		AcceptedAssets: "USDT",
		Payload:        fmt.Sprintf("purchaseId=%d&username=%s", purchaseID, ctx.Value(contextkey.Username)),
		Description:    fmt.Sprintf("Subscription on %d month", months),
		PaidBtnName:    "callback",
		PaidBtnUrl:     config.BotURL(),
	})
	if err != nil {
		slog.Error("Error creating invoice", "err", err)
		return "", 0, err
	}

	updates := map[string]interface{}{
		"crypto_invoice_url": invoice.BotInvoiceUrl,
		"crypto_invoice_id":  invoice.InvoiceID,
		"status":             domainpurchase.StatusPending,
	}

	if err = p.repo.UpdateFields(ctx, purchaseID, updates); err != nil {
		slog.Error("Error updating purchase", "err", err)
		return "", 0, err
	}

	return invoice.BotInvoiceUrl, purchaseID, nil
}
