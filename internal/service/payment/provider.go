package payment

import (
	"context"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	domainpurchase "remnawave-tg-shop-bot/internal/domain/purchase"
)

// Provider describes payment provider behaviour.
type Provider interface {
	// Type returns invoice type handled by provider.
	Type() domainpurchase.InvoiceType
	// Enabled reports whether provider is available.
	Enabled() bool
	// CreateInvoice creates a new purchase and returns payment URL and purchase ID.
	CreateInvoice(ctx context.Context, amount int, months int, customer *domaincustomer.Customer) (string, int64, error)
}
