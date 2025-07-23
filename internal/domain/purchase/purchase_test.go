package purchase

import "testing"

func TestInvoiceTypeValues(t *testing.T) {
	if InvoiceTypeCrypto == "" || InvoiceTypeTelegram == "" || InvoiceTypeTribute == "" {
		t.Fatal("invoice types should not be empty")
	}
}

func TestStatusValues(t *testing.T) {
	if StatusNew == "" || StatusPending == "" || StatusPaid == "" || StatusCancel == "" {
		t.Fatal("status constants empty")
	}
}
