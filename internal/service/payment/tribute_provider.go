package payment

import (
	"context"
	"log/slog"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	domainpurchase "remnawave-tg-shop-bot/internal/domain/purchase"
	"remnawave-tg-shop-bot/internal/pkg/config"
)

// TributeProvider implements Provider for Tribute payments.
type TributeProvider struct {
	repo PurchaseRepository
}

func NewTributeProvider(repo PurchaseRepository) *TributeProvider {
	return &TributeProvider{repo: repo}
}

func (p TributeProvider) Type() domainpurchase.InvoiceType { return domainpurchase.InvoiceTypeTribute }

func (p TributeProvider) Enabled() bool { return config.GetTributePaymentUrl() != "" }

func (p TributeProvider) CreateInvoice(ctx context.Context, amount int, months int, customer *domaincustomer.Customer) (string, int64, error) {
	purchaseID, err := p.repo.Create(ctx, &domainpurchase.Purchase{
		InvoiceType: domainpurchase.InvoiceTypeTribute,
		Status:      domainpurchase.StatusPending,
		Amount:      float64(amount),
		Currency:    "RUB",
		CustomerID:  customer.ID,
		Month:       months,
	})
	if err != nil {
		slog.Error("Error creating purchase", "err", err)
		return "", 0, err
	}

	return config.GetTributePaymentUrl(), purchaseID, nil
}
