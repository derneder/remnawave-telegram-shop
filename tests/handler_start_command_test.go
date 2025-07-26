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
	"remnawave-tg-shop-bot/internal/pkg/translation"
	referralrepo "remnawave-tg-shop-bot/internal/repository/referral"
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

func TestStartCommandHandler_ReferralSaved(t *testing.T) {
	trans := translation.GetInstance()
	if err := trans.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}

	custRepo := &customerRepoNotFound{}
	refRepo := &StubReferralRepo{}
	h := handlerpkg.NewHandler(nil, nil, trans, custRepo, nil, refRepo, nil, nil, nil, nil)

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

	if refRepo.CreatedReferrerID != 5 || refRepo.CreatedRefereeID != 2 {
		t.Fatalf("referral not created")
	}
}

func TestStartCommandHandler_ReferralSelf(t *testing.T) {
	trans := translation.GetInstance()
	if err := trans.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}

	custRepo := &customerRepoNotFound{}
	refRepo := &StubReferralRepo{}
	h := handlerpkg.NewHandler(nil, nil, trans, custRepo, nil, refRepo, nil, nil, nil, nil)

	b, _ := bot.New("token", bot.WithHTTPClient(time.Second, &startHTTPClient{}), bot.WithSkipGetMe())

	upd := &models.Update{Message: &models.Message{Chat: models.Chat{ID: 2}, From: &models.User{ID: 2, LanguageCode: "en", FirstName: "u"}, Text: "/start ref_2", Entities: []models.MessageEntity{{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: len("/start")}}}}

	h.StartCommandHandler(context.Background(), b, upd)

	if refRepo.CreatedReferrerID != 0 {
		t.Fatalf("self referral should not be saved")
	}
}

func TestStartCommandHandler_ReferralDuplicate(t *testing.T) {
	trans := translation.GetInstance()
	if err := trans.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}

	repo := &StubCustomerRepo{}
	refRepo := &StubReferralRepo{Model: &referralrepo.Model{RefereeID: 3}}
	h := handlerpkg.NewHandler(nil, nil, trans, repo, nil, refRepo, nil, nil, nil, nil)

	b, _ := bot.New("token", bot.WithHTTPClient(time.Second, &startHTTPClient{}), bot.WithSkipGetMe())

	upd := &models.Update{Message: &models.Message{Chat: models.Chat{ID: 3}, From: &models.User{ID: 3, LanguageCode: "en", FirstName: "u"}, Text: "/start ref_5", Entities: []models.MessageEntity{{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: len("/start")}}}}

	h.StartCommandHandler(context.Background(), b, upd)

	if refRepo.CreatedReferrerID != 0 {
		t.Fatalf("duplicate referral saved")
	}
}
