package purchase

import (
	"time"
)

type InvoiceType string

const (
	InvoiceTypeCrypto   InvoiceType = "crypto"
	InvoiceTypeTelegram InvoiceType = "telegram"
	InvoiceTypeTribute  InvoiceType = "tribute"
)

type Status string

const (
	StatusNew     Status = "new"
	StatusPending Status = "pending"
	StatusPaid    Status = "paid"
	StatusCancel  Status = "cancel"
)

type Purchase struct {
	ID                int64
	Amount            float64
	CustomerID        int64
	CreatedAt         time.Time
	Month             int
	PaidAt            *time.Time
	Currency          string
	ExpireAt          *time.Time
	Status            Status
	InvoiceType       InvoiceType
	CryptoInvoiceID   *int64
	CryptoInvoiceLink *string
}
