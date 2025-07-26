package handler

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnawave-tg-shop-bot/internal/pkg/translation"
	pg "remnawave-tg-shop-bot/internal/repository/pg"
	uimenu "remnawave-tg-shop-bot/internal/ui/menu"
	testutils "remnawave-tg-shop-bot/tests"
)

type stubPromoRepo struct{ promos []pg.Promocode }

func (s *stubPromoRepo) FindByCreator(ctx context.Context, id int64) ([]pg.Promocode, error) {
	return s.promos, nil
}

type stubHTTPPromo struct{ body string }

func (h *stubHTTPPromo) Do(r *http.Request) (*http.Response, error) {
	b, _ := io.ReadAll(r.Body)
	h.body = string(b)
	resp := &http.Response{StatusCode: http.StatusOK}
	resp.Body = io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`))
	return resp, nil
}

func TestPromoMyListCallbackHandler_Filter(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatal(err)
	}

	repo := &stubPromoRepo{promos: []pg.Promocode{
		{ID: 1, Code: "A1", UsesLeft: 1, CreatedBy: 1, Active: true},
		{ID: 2, Code: "B1", UsesLeft: 0, CreatedBy: 1, Active: false},
		{ID: 3, Code: "C1", UsesLeft: 0, CreatedBy: 1, Active: true, Deleted: true},
		{ID: 4, Code: "D1", UsesLeft: -1, CreatedBy: 1, Active: true},
	}}

	httpc := &stubHTTPPromo{}
	b, _ := bot.New("t", bot.WithHTTPClient(time.Second, httpc), bot.WithSkipGetMe())

	h := NewHandler(nil, nil, tm, &testutils.StubCustomerRepo{}, nil, nil, repo, nil, nil, nil)

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{ID: "1", From: models.User{ID: 1, LanguageCode: "en"}, Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}}}}
	upd.CallbackQuery.Data = uimenu.CallbackPromoMyList

	h.PromoMyListCallbackHandler(context.Background(), b, upd)

	if strings.Contains(httpc.body, "C1") || strings.Contains(httpc.body, "D1") {
		t.Fatalf("unexpected promo codes in list: %s", httpc.body)
	}
	if !strings.Contains(httpc.body, "A1") || !strings.Contains(httpc.body, "B1") {
		t.Fatalf("expected promo codes missing: %s", httpc.body)
	}
}
