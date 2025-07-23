package tests

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	handlerpkg "remnawave-tg-shop-bot/internal/adapter/telegram/handler"
	domainpurchase "remnawave-tg-shop-bot/internal/domain/purchase"
	"remnawave-tg-shop-bot/internal/pkg/cache"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	"remnawave-tg-shop-bot/internal/service/payment"
)

// stub implementations

type stubPurchaseRepo struct {
	ctxCreate context.Context
	ctxUpdate context.Context
}

func (s *stubPurchaseRepo) Create(ctx context.Context, p *domainpurchase.Purchase) (int64, error) {
	s.ctxCreate = ctx
	return 1, nil
}
func (s *stubPurchaseRepo) FindByInvoiceTypeAndStatus(ctx context.Context, invoiceType domainpurchase.InvoiceType, status domainpurchase.Status) (*[]domainpurchase.Purchase, error) {
	return nil, nil
}
func (s *stubPurchaseRepo) FindById(ctx context.Context, id int64) (*domainpurchase.Purchase, error) {
	return nil, nil
}
func (s *stubPurchaseRepo) UpdateFields(ctx context.Context, id int64, updates map[string]interface{}) error {
	s.ctxUpdate = ctx
	return nil
}
func (s *stubPurchaseRepo) MarkAsPaid(ctx context.Context, purchaseID int64) error { return nil }

type stubMessenger struct{ ctx context.Context }

func (m *stubMessenger) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	m.ctx = ctx
	return &models.Message{}, nil
}
func (m *stubMessenger) DeleteMessage(ctx context.Context, params *bot.DeleteMessageParams) (bool, error) {
	m.ctx = ctx
	return true, nil
}
func (m *stubMessenger) CreateInvoiceLink(ctx context.Context, params *bot.CreateInvoiceLinkParams) (string, error) {
	m.ctx = ctx
	return "link", nil
}

type httpClient struct{}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	resp := &http.Response{StatusCode: http.StatusOK}
	resp.Body = io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{"message_id":1}}`)))
	return resp, nil
}

func TestPaymentCallbackHandler_ContextPropagation(t *testing.T) {
	custRepo := &StubCustomerRepo{}
	purchRepo := &stubPurchaseRepo{}
	messenger := &stubMessenger{}
	cache := cache.NewCache(context.Background(), time.Minute)
	defer cache.Close()
	trans := translation.GetInstance()
	paySvc := payment.NewPaymentService(trans, purchRepo, nil, custRepo, messenger, nil, nil, nil, nil, cache)

	h := handlerpkg.NewHandler(nil, paySvc, trans, custRepo, nil, nil, nil, nil, cache)

	b, err := bot.New("token", bot.WithHTTPClient(time.Second, &httpClient{}), bot.WithSkipGetMe())
	if err != nil {
		t.Fatalf("bot init: %v", err)
	}

	upd := &models.Update{
		CallbackQuery: &models.CallbackQuery{
			Data:    "payment?month=1&amount=10&invoiceType=telegram",
			From:    models.User{ID: 1, LanguageCode: "en", Username: "user"},
			Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}},
		},
	}

	ctx := context.WithValue(context.Background(), CtxKey{}, "v")
	h.PaymentCallbackHandler(ctx, b, upd)

	if custRepo.Ctx.Value(CtxKey{}) != "v" {
		t.Errorf("context not propagated to repository")
	}
	if purchRepo.ctxCreate.Value(CtxKey{}) != "v" || purchRepo.ctxUpdate.Value(CtxKey{}) != "v" {
		t.Errorf("context not propagated to purchase repository")
	}
	if messenger.ctx.Value(CtxKey{}) != "v" {
		t.Errorf("context not propagated to messenger")
	}
}
