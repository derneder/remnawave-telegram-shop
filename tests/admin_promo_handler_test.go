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
	"remnawave-tg-shop-bot/internal/pkg/contextkey"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	uimenu "remnawave-tg-shop-bot/internal/ui/menu"
)

type promoServiceStub struct {
	bal struct{ amount, limit int }
	sub struct {
		code        string
		days, limit int
	}
}

func (s *promoServiceStub) CreateSubscription(ctx context.Context, code string, days, limit int, by int64) error {
	s.sub = struct {
		code        string
		days, limit int
	}{code, days, limit}
	return nil
}

func (s *promoServiceStub) CreateBalance(ctx context.Context, amount, limit int, by int64) (string, error) {
	s.bal = struct{ amount, limit int }{amount, limit}
	return "CODE", nil
}

type stubHTTP struct{ body string }

func (h *stubHTTP) Do(r *http.Request) (*http.Response, error) {
	b, _ := io.ReadAll(r.Body)
	h.body = string(b)
	resp := &http.Response{StatusCode: http.StatusOK}
	resp.Body = io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`))
	return resp, nil
}

func TestAdminPromoBalanceWizard(t *testing.T) {
	SetTestEnv(t)
	tm := translation.GetInstance()
	_ = tm.InitDefaultTranslations()
	svc := &promoServiceStub{}
	httpc := &stubHTTP{}
	b, _ := bot.New("t", bot.WithHTTPClient(time.Second, httpc), bot.WithSkipGetMe())
	h := handlerpkg.NewHandler(nil, nil, tm, &StubCustomerRepo{}, nil, nil, nil, nil, svc, nil)

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{ID: "1", From: models.User{ID: 1, LanguageCode: "ru"}, Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}}}}

	ctx := context.WithValue(context.Background(), contextkey.IsAdminKey, true)
	upd.CallbackQuery.Data = uimenu.CallbackPromoAdminMenu
	h.AdminPromoCallbackHandler(ctx, b, upd)

	upd.CallbackQuery.Data = uimenu.CallbackPromoAdminBalanceStart
	h.AdminPromoCallbackHandler(ctx, b, upd)

	upd.CallbackQuery.Data = uimenu.CallbackPromoAdminBalanceAmount + ":100"
	h.AdminPromoCallbackHandler(ctx, b, upd)

	upd.CallbackQuery.Data = uimenu.CallbackPromoAdminBalanceLimit + ":1"
	h.AdminPromoCallbackHandler(ctx, b, upd)

	upd.CallbackQuery.Data = uimenu.CallbackPromoAdminBalanceConfirm
	h.AdminPromoCallbackHandler(ctx, b, upd)

	if svc.bal.amount != 10000 || svc.bal.limit != 1 {
		t.Fatalf("unexpected svc args %#v", svc.bal)
	}
}

func TestAdminPromoSubWizard(t *testing.T) {
	SetTestEnv(t)
	tm := translation.GetInstance()
	_ = tm.InitDefaultTranslations()
	svc := &promoServiceStub{}
	httpc := &stubHTTP{}
	b, _ := bot.New("t", bot.WithHTTPClient(time.Second, httpc), bot.WithSkipGetMe())
	h := handlerpkg.NewHandler(nil, nil, tm, &StubCustomerRepo{}, nil, nil, nil, nil, svc, nil)

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{ID: "1", From: models.User{ID: 1, LanguageCode: "ru"}, Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}}}}
	ctx := context.WithValue(context.Background(), contextkey.IsAdminKey, true)

	upd.CallbackQuery.Data = uimenu.CallbackPromoAdminMenu
	h.AdminPromoCallbackHandler(ctx, b, upd)

	upd.CallbackQuery.Data = uimenu.CallbackPromoAdminSubStart
	h.AdminPromoCallbackHandler(ctx, b, upd)

	upd.CallbackQuery.Data = uimenu.CallbackPromoAdminSubCodeRandom
	h.AdminPromoCallbackHandler(ctx, b, upd)

	upd.CallbackQuery.Data = uimenu.CallbackPromoAdminSubDays + ":30"
	h.AdminPromoCallbackHandler(ctx, b, upd)

	upd.CallbackQuery.Data = uimenu.CallbackPromoAdminSubLimit + ":2"
	h.AdminPromoCallbackHandler(ctx, b, upd)

	upd.CallbackQuery.Data = uimenu.CallbackPromoAdminSubConfirm
	h.AdminPromoCallbackHandler(ctx, b, upd)

	if svc.sub.days != 30 || svc.sub.limit != 2 {
		t.Fatalf("unexpected sub args %#v", svc.sub)
	}
}
