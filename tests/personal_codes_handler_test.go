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
	"remnawave-tg-shop-bot/internal/pkg/translation"
	uimenu "remnawave-tg-shop-bot/internal/ui/menu"
)

type stubHTTPPersonal struct{ body string }

func (h *stubHTTPPersonal) Do(r *http.Request) (*http.Response, error) {
	b, _ := io.ReadAll(r.Body)
	h.body = string(b)
	resp := &http.Response{StatusCode: http.StatusOK}
	resp.Body = io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`))
	return resp, nil
}

func TestPersonalCodesMenu(t *testing.T) {
	SetTestEnv(t)
	tm := translation.GetInstance()
	_ = tm.InitDefaultTranslations()
	httpc := &stubHTTPPersonal{}
	b, _ := bot.New("t", bot.WithHTTPClient(time.Second, httpc), bot.WithSkipGetMe())
	h := handlerpkg.NewHandler(nil, nil, tm, &StubCustomerRepo{}, nil, nil, nil, nil, nil, nil)

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{ID: "1", From: models.User{ID: 1, LanguageCode: "ru"}, Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}}}}
	upd.CallbackQuery.Data = uimenu.CallbackPersonalCodes

	h.PersonalCodesCallbackHandler(context.Background(), b, upd)

	if !strings.Contains(httpc.body, tm.GetText("ru", "personal_create_button")) {
		t.Fatalf("menu not sent")
	}
}

func TestPersonalCreateFlow(t *testing.T) {
	SetTestEnv(t)
	tm := translation.GetInstance()
	_ = tm.InitDefaultTranslations()
	httpc := &stubHTTPPersonal{}
	b, _ := bot.New("t", bot.WithHTTPClient(time.Second, httpc), bot.WithSkipGetMe())
	h := handlerpkg.NewHandler(nil, nil, tm, &StubCustomerRepo{}, nil, nil, nil, nil, nil, nil)

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{ID: "1", From: models.User{ID: 1, LanguageCode: "ru"}, Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}}}}
	upd.CallbackQuery.Data = uimenu.CallbackPersonalCreate
	h.PersonalCreateCallbackHandler(context.Background(), b, upd)
	if !strings.Contains(httpc.body, tm.GetText("ru", "personal_months_prompt")) {
		t.Fatalf("months prompt not sent")
	}

	msgUpd := &models.Update{Message: &models.Message{ID: 10, Chat: models.Chat{ID: 1}, From: &models.User{ID: 1, LanguageCode: "ru"}, Text: "1"}}
	h.PersonalMessageHandler(context.Background(), b, msgUpd)
	if !strings.Contains(httpc.body, tm.GetText("ru", "personal_uses_prompt")) {
		t.Fatalf("uses prompt not sent")
	}

	msgUpd.Message.Text = "2"
	h.PersonalMessageHandler(context.Background(), b, msgUpd)
}
