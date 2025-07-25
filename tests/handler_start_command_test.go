package tests

import (
	"context"
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
	"remnawave-tg-shop-bot/internal/pkg/translation"
	"remnawave-tg-shop-bot/internal/service/payment"
)

type startHTTPClient struct{}

func (c *startHTTPClient) Do(req *http.Request) (*http.Response, error) {
	resp := &http.Response{StatusCode: http.StatusOK}
	resp.Body = io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`))
	return resp, nil
}

func TestStartCommandHandler_NoArgs(t *testing.T) {
	trans := translation.GetInstance()
	if err := trans.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}

	repo := &StubCustomerRepo{}
       h := handlerpkg.NewHandler(nil, nil, trans, repo, nil, nil, nil, nil, nil, nil)

	b, err := bot.New("token", bot.WithHTTPClient(time.Second, &startHTTPClient{}), bot.WithSkipGetMe())
	if err != nil {
		t.Fatalf("bot init: %v", err)
	}

	upd := &models.Update{
		Message: &models.Message{
			Chat:     models.Chat{ID: 1},
			From:     &models.User{ID: 1, LanguageCode: "en", FirstName: "u"},
			Text:     "/start",
			Entities: []models.MessageEntity{{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: len("/start")}},
		},
	}

	ctx := context.WithValue(context.Background(), CtxKey{}, "v")
	h.StartCommandHandler(ctx, b, upd)

	if repo.Ctx.Value(CtxKey{}) != "v" {
		t.Errorf("context not propagated")
	}
}

// customerRepoNotFound always returns nil on FindByTelegramId to simulate a new user.
type customerRepoNotFound struct{ StubCustomerRepo }

func (c *customerRepoNotFound) FindByTelegramId(ctx context.Context, id int64) (*domaincustomer.Customer, error) {
	c.Ctx = ctx
	c.Calls++
	if id == 2 {
		return nil, nil
	}
	return &domaincustomer.Customer{ID: 1, TelegramID: id, Language: "en"}, nil
}
func (c *customerRepoNotFound) FindById(ctx context.Context, id int64) (*domaincustomer.Customer, error) {
	return &domaincustomer.Customer{ID: id, TelegramID: id, Language: "en"}, nil
}

type purchaseRepoStub struct{ stubPurchaseRepoSimple }

func (purchaseRepoStub) FindById(ctx context.Context, id int64) (*domainpurchase.Purchase, error) {
	return &domainpurchase.Purchase{ID: id, CustomerID: id, Amount: 10, Status: domainpurchase.StatusNew}, nil
}

func TestStartCommandHandler_ReferralMarksGranted(t *testing.T) {
	trans := translation.GetInstance()
	if err := trans.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}

	custRepo := &customerRepoNotFound{}
	refRepo := &StubReferralRepo{}
	h := handlerpkg.NewHandler(nil, nil, trans, custRepo, nil, refRepo, nil, nil, nil)

	b, err := bot.New("token", bot.WithHTTPClient(time.Second, &startHTTPClient{}), bot.WithSkipGetMe())
	if err != nil {
		t.Fatalf("bot init: %v", err)
	}

	upd := &models.Update{
		Message: &models.Message{
			Chat:     models.Chat{ID: 2},
			From:     &models.User{ID: 2, LanguageCode: "en", FirstName: "u"},
			Text:     "/start ref_5",
			Entities: []models.MessageEntity{{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: len("/start")}},
		},
	}

	h.StartCommandHandler(context.Background(), b, upd)

	if refRepo.MarkedID != 1 {
		t.Fatalf("bonus not marked granted")
	}

	// ProcessPurchaseById should not grant bonus again when already marked.
	purchRepo := purchaseRepoStub{}
	c := cache.NewCache(context.Background(), time.Minute)
	defer c.Close()
	paySvc := payment.NewPaymentService(trans, purchRepo, nil, custRepo, &stubMessenger{}, nil, refRepo, nil, nil, c)

	if err := paySvc.ProcessPurchaseById(context.Background(), 0); err != nil {
		t.Fatalf("process purchase: %v", err)
	}

	if refRepo.MarkedID != 1 {
		t.Errorf("bonus marked again unexpectedly")
	}
}
