package payment_test

import (
	"context"
	"testing"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	domainpurchase "remnawave-tg-shop-bot/internal/domain/purchase"
	"remnawave-tg-shop-bot/internal/service/payment"
)

type stubProvider struct {
	typ     domainpurchase.InvoiceType
	enabled bool
	called  bool
}

func (s *stubProvider) Type() domainpurchase.InvoiceType { return s.typ }
func (s *stubProvider) Enabled() bool                    { return s.enabled }
func (s *stubProvider) CreateInvoice(ctx context.Context, amount int, months int, c *domaincustomer.Customer) (string, int64, error) {
	s.called = true
	return "url", 1, nil
}

type stubRepo struct{}

func (stubRepo) Create(ctx context.Context, p *domainpurchase.Purchase) (int64, error) { return 1, nil }
func (stubRepo) FindByInvoiceTypeAndStatus(ctx context.Context, t domainpurchase.InvoiceType, s domainpurchase.Status) (*[]domainpurchase.Purchase, error) {
	return nil, nil
}
func (stubRepo) FindById(ctx context.Context, id int64) (*domainpurchase.Purchase, error) {
	return nil, nil
}
func (stubRepo) UpdateFields(ctx context.Context, id int64, m map[string]interface{}) error {
	return nil
}
func (stubRepo) MarkAsPaid(ctx context.Context, id int64) error { return nil }

func TestEnabledProviders(t *testing.T) {
	p1 := &stubProvider{typ: domainpurchase.InvoiceTypeCrypto, enabled: true}
	p2 := &stubProvider{typ: domainpurchase.InvoiceTypeTribute, enabled: false}
	svc := payment.NewPaymentService(nil, nil, nil, nil, nil, []payment.Provider{p1, p2}, nil, nil, nil, nil)
	res := svc.EnabledProviders()
	if len(res) != 1 || res[0] != p1 {
		t.Fatalf("expected only enabled provider")
	}
}

func TestCreatePurchaseUnknownType(t *testing.T) {
	svc := payment.NewPaymentService(nil, stubRepo{}, nil, nil, nil, nil, nil, nil, nil, nil)
	c := &domaincustomer.Customer{ID: 1}
	if _, _, err := svc.CreatePurchase(context.Background(), 10, 1, c, domainpurchase.InvoiceTypeCrypto); err == nil {
		t.Fatal("expected error")
	}
}
