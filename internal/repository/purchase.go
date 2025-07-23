package repository

import (
	"context"
	"remnawave-tg-shop-bot/internal/domain/purchase"
)

type PurchaseRepository interface {
	Create(ctx context.Context, p *purchase.Purchase) (int64, error)
	FindByInvoiceTypeAndStatus(ctx context.Context, invoiceType purchase.InvoiceType, status purchase.Status) (*[]purchase.Purchase, error)
	FindById(ctx context.Context, id int64) (*purchase.Purchase, error)
	UpdateFields(ctx context.Context, id int64, updates map[string]interface{}) error
	MarkAsPaid(ctx context.Context, purchaseID int64) error
}
