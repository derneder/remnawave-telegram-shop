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
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

type stubPromoService struct {
	sub struct {
		code        string
		days, limit int
		by          int64
	}
	bal struct {
		amount, limit int
		by            int64
	}
}

type promoHTTPClient struct{}

func (c *promoHTTPClient) Do(req *http.Request) (*http.Response, error) {
	resp := &http.Response{StatusCode: http.StatusOK}
	resp.Body = io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{"message_id":1}}`)))
	return resp, nil
}

func (s *stubPromoService) CreateSubscription(ctx context.Context, code string, days, limit int, by int64) error {
	s.sub = struct {
		code        string
		days, limit int
		by          int64
	}{code, days, limit, by}
	return nil
}

func (s *stubPromoService) CreateBalance(ctx context.Context, amount, limit int, by int64) (string, error) {
	s.bal = struct {
		amount, limit int
		by            int64
	}{amount, limit, by}
	return "CODE12345678901234", nil
}

func TestAddSubPromoCommandHandler(t *testing.T) {
	trans := translation.GetInstance()
	_ = trans.InitDefaultTranslations()
	svc := &stubPromoService{}
	h := handlerpkg.NewHandler(nil, nil, trans, &StubCustomerRepo{}, nil, nil, nil, nil, svc, nil)
	b, _ := bot.New("t", bot.WithSkipGetMe(), bot.WithHTTPClient(time.Second, &promoHTTPClient{}))
	upd := &models.Update{Message: &models.Message{Chat: models.Chat{ID: 1}, Text: "/addsubpromo ABC 30 2", From: &models.User{ID: 1}}}
	h.AddSubPromoCommandHandler(context.Background(), b, upd)
	if svc.sub.code != "ABC" || svc.sub.days != 30 || svc.sub.limit != 2 || svc.sub.by != 1 {
		t.Fatalf("wrong args %#v", svc.sub)
	}
	upd.Message.Text = "/addsubpromo bad"
	svc.sub = struct {
		code        string
		days, limit int
		by          int64
	}{}
	h.AddSubPromoCommandHandler(context.Background(), b, upd)
	if svc.sub.code != "" {
		t.Fatal("called on invalid args")
	}
}

func TestAddBalPromoCommandHandler(t *testing.T) {
	trans := translation.GetInstance()
	_ = trans.InitDefaultTranslations()
	svc := &stubPromoService{}
	h := handlerpkg.NewHandler(nil, nil, trans, &StubCustomerRepo{}, nil, nil, nil, nil, svc, nil)
	b, _ := bot.New("t", bot.WithSkipGetMe(), bot.WithHTTPClient(time.Second, &promoHTTPClient{}))
	upd := &models.Update{Message: &models.Message{Chat: models.Chat{ID: 1}, Text: "/addbalpromo 100 3", From: &models.User{ID: 1}}}
	h.AddBalPromoCommandHandler(context.Background(), b, upd)
	if svc.bal.amount != 10000 || svc.bal.limit != 3 || svc.bal.by != 1 {
		t.Fatalf("wrong args %#v", svc.bal)
	}
	upd.Message.Text = "/addbalpromo bad"
	svc.bal = struct {
		amount, limit int
		by            int64
	}{}
	h.AddBalPromoCommandHandler(context.Background(), b, upd)
	if svc.bal.amount != 0 {
		t.Fatal("called on invalid args")
	}
}
