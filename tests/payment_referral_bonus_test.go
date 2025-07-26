package tests

import (
	"context"
	"testing"
	"time"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	domainpurchase "remnawave-tg-shop-bot/internal/domain/purchase"
	"remnawave-tg-shop-bot/internal/pkg/cache"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	referralrepo "remnawave-tg-shop-bot/internal/repository/referral"
	"remnawave-tg-shop-bot/internal/service/payment"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// stub implementations for this test

type stubPurchaseRepoBonus struct{ purchase *domainpurchase.Purchase }

func (s *stubPurchaseRepoBonus) Create(ctx context.Context, p *domainpurchase.Purchase) (int64, error) {
	return 0, nil
}
func (s *stubPurchaseRepoBonus) FindByInvoiceTypeAndStatus(ctx context.Context, t domainpurchase.InvoiceType, st domainpurchase.Status) (*[]domainpurchase.Purchase, error) {
	return nil, nil
}
func (s *stubPurchaseRepoBonus) FindById(ctx context.Context, id int64) (*domainpurchase.Purchase, error) {
	return s.purchase, nil
}
func (s *stubPurchaseRepoBonus) UpdateFields(ctx context.Context, id int64, updates map[string]interface{}) error {
	return nil
}
func (s *stubPurchaseRepoBonus) MarkAsPaid(ctx context.Context, id int64) error { return nil }

// customer repo tracking balance updates

type stubCustomerRepoBonus struct {
	customers   map[int64]*domaincustomer.Customer
	updateCount map[int64]int
}

func (r *stubCustomerRepoBonus) FindById(ctx context.Context, id int64) (*domaincustomer.Customer, error) {
	if c, ok := r.customers[id]; ok {
		return c, nil
	}
	return nil, nil
}
func (r *stubCustomerRepoBonus) FindByTelegramId(ctx context.Context, tgID int64) (*domaincustomer.Customer, error) {
	for _, c := range r.customers {
		if c.TelegramID == tgID {
			return c, nil
		}
	}
	return nil, nil
}
func (r *stubCustomerRepoBonus) Create(ctx context.Context, c *domaincustomer.Customer) (*domaincustomer.Customer, error) {
	return c, nil
}
func (r *stubCustomerRepoBonus) UpdateFields(ctx context.Context, id int64, updates map[string]interface{}) error {
	if r.updateCount == nil {
		r.updateCount = make(map[int64]int)
	}
	r.updateCount[id]++
	if c, ok := r.customers[id]; ok {
		if b, ok2 := updates["balance"].(float64); ok2 {
			c.Balance = b
		}
	}
	return nil
}
func (r *stubCustomerRepoBonus) FindByTelegramIds(ctx context.Context, ids []int64) ([]domaincustomer.Customer, error) {
	return nil, nil
}
func (r *stubCustomerRepoBonus) DeleteByNotInTelegramIds(ctx context.Context, ids []int64) error {
	return nil
}
func (r *stubCustomerRepoBonus) CreateBatch(ctx context.Context, customers []domaincustomer.Customer) error {
	return nil
}
func (r *stubCustomerRepoBonus) UpdateBatch(ctx context.Context, customers []domaincustomer.Customer) error {
	return nil
}
func (r *stubCustomerRepoBonus) FindByExpirationRange(ctx context.Context, start, end time.Time) (*[]domaincustomer.Customer, error) {
	return nil, nil
}

// messenger stub recording messages

type stubMessengerBonus struct{ texts []string }

func (m *stubMessengerBonus) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	m.texts = append(m.texts, params.Text)
	return &models.Message{}, nil
}
func (m *stubMessengerBonus) DeleteMessage(ctx context.Context, params *bot.DeleteMessageParams) (bool, error) {
	return true, nil
}
func (m *stubMessengerBonus) CreateInvoiceLink(ctx context.Context, params *bot.CreateInvoiceLinkParams) (string, error) {
	return "", nil
}

func TestProcessPurchaseById_ReferralBonusOnce(t *testing.T) {
	SetTestEnv(t)
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatal(err)
	}

	purchRepo := &stubPurchaseRepoBonus{purchase: &domainpurchase.Purchase{ID: 1, Amount: 10, CustomerID: 1}}
	customers := map[int64]*domaincustomer.Customer{
		1: {ID: 1, TelegramID: 1, Language: "en", Balance: 0},
		2: {ID: 2, TelegramID: 2, Language: "en", Balance: 0},
	}
	custRepo := &stubCustomerRepoBonus{customers: customers}
	refModel := &referralrepo.Model{ID: 1, ReferrerID: 2, RefereeID: 1, CreatedAt: time.Now(), BonusGranted: false}
	refRepo := &StubReferralRepo{Model: refModel}
	messenger := &stubMessengerBonus{}
	cache := cache.NewCache(context.Background(), time.Minute)
	defer cache.Close()

	svc := payment.NewPaymentService(tm, purchRepo, nil, custRepo, messenger, nil, refRepo, nil, nil, cache)

	if err := svc.ProcessPurchaseById(context.Background(), 1); err != nil {
		t.Fatalf("first: %v", err)
	}
	if !refRepo.Model.BonusGranted {
		t.Fatalf("bonus not marked")
	}
	if custRepo.updateCount[2] != 1 {
		t.Fatalf("referrer not updated")
	}
	if len(messenger.texts) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messenger.texts))
	}

	purchRepo.purchase = &domainpurchase.Purchase{ID: 2, Amount: 5, CustomerID: 1}
	if err := svc.ProcessPurchaseById(context.Background(), 2); err != nil {
		t.Fatalf("second: %v", err)
	}
	if custRepo.updateCount[2] != 1 {
		t.Fatalf("bonus granted twice")
	}
	if len(messenger.texts) != 3 {
		t.Fatalf("unexpected message count: %d", len(messenger.texts))
	}
}
