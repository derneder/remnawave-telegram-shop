package tests

import (
	"testing"

	"remnawave-tg-shop-bot/internal/domain/purchase"
)

func TestInvoiceTypeValues(t *testing.T) {
	if purchase.InvoiceTypeCrypto == "" || purchase.InvoiceTypeTelegram == "" || purchase.InvoiceTypeTribute == "" {
		t.Fatal("invoice types should not be empty")
	}
}

func TestStatusValues(t *testing.T) {
	if purchase.StatusNew == "" || purchase.StatusPending == "" || purchase.StatusPaid == "" || purchase.StatusCancel == "" {
		t.Fatal("status constants empty")
	}
}
