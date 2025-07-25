package tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	handlerpkg "remnawave-tg-shop-bot/internal/adapter/telegram/handler"
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	domainpurchase "remnawave-tg-shop-bot/internal/domain/purchase"
	"remnawave-tg-shop-bot/internal/pkg/cache"
	"remnawave-tg-shop-bot/internal/pkg/config"
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

type bodyClient struct{ body []byte }

func (c *bodyClient) Do(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		c.body, _ = io.ReadAll(req.Body)
	}
	resp := &http.Response{StatusCode: http.StatusOK}
	resp.Body = io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`))
	return resp, nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

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

func TestSellCallbackHandler_DynamicText(t *testing.T) {
	SetTestEnv(t)
	trans := translation.GetInstance()
	if err := trans.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}

	config.InitConfig()
	repo := &StubCustomerRepo{CustomerByTelegramID: &domaincustomer.Customer{TelegramID: 1, Balance: 10, Language: "en"}}
	h := handlerpkg.NewHandler(nil, nil, trans, repo, nil, nil, nil, nil, nil)

	client := &bodyClient{}
	b, err := bot.New("token", bot.WithHTTPClient(time.Second, client), bot.WithSkipGetMe())
	if err != nil {
		t.Fatalf("bot init: %v", err)
	}

	upd := &models.Update{
		CallbackQuery: &models.CallbackQuery{
			Data:    "sell?month=1&amount=1",
			From:    models.User{ID: 1, LanguageCode: "en"},
			Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}},
		},
	}

	h.SellCallbackHandler(context.Background(), b, upd)

	body := string(client.body)
	expected := trans.GetText("en", "choose_plan_header") + "\n\n" +
		fmt.Sprintf(trans.GetText("en", "choose_plan_balance"), 10) + "\n\n" +
		fmt.Sprintf(trans.GetText("en", "choose_plan_line"), trans.GetText("en", "month_1"), config.Price1()) + "\n\n" +
		trans.GetText("en", "choose_plan_footer")
	if !strings.Contains(body, expected) {
		t.Fatalf("text not found in body: %s", body)
	}
	if !strings.Contains(body, handlerpkg.CallbackPayFromBal+"?month=1") {
		t.Fatalf("callback not found")
	}
}

func TestQRCallbackHandler_MiniAppURL(t *testing.T) {
	SetTestEnv(t)
	t.Setenv("MINI_APP_URL", "https://app")
	trans := translation.GetInstance()
	if err := trans.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}

	config.InitConfig()

	link := "https://sub"
	repo := &StubCustomerRepo{CustomerByTelegramID: &domaincustomer.Customer{TelegramID: 1, Language: "ru", SubscriptionLink: &link}}
	h := handlerpkg.NewHandler(nil, nil, trans, repo, nil, nil, nil, nil, nil)

	client := &bodyClient{}
	b, err := bot.New("token", bot.WithHTTPClient(time.Second, client), bot.WithSkipGetMe())
	if err != nil {
		t.Fatalf("bot init: %v", err)
	}

	oldTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("data")), Header: make(http.Header)}, nil
	})
	defer func() { http.DefaultTransport = oldTransport }()

	upd := &models.Update{
		CallbackQuery: &models.CallbackQuery{
			Data:    "qr",
			From:    models.User{ID: 1, LanguageCode: "ru"},
			Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}},
		},
	}

	h.QRCallbackHandler(context.Background(), b, upd)

	body := string(client.body)
	t.Logf("body: %s", body)
	expectedCaption := fmt.Sprintf(trans.GetText("ru", "qr_text"), "")
	if !strings.Contains(body, expectedCaption) {
		t.Fatalf("caption not found")
	}
	if !strings.Contains(body, "https://app") {
		t.Fatalf("url not found")
	}
}
