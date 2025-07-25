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
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

type qrRoundTripper struct{ bodies []string }

func (t *qrRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var b []byte
	if req.Body != nil {
		b, _ = io.ReadAll(req.Body)
		req.Body.Close()
	}
	t.bodies = append(t.bodies, string(b))
	if strings.Contains(req.URL.Host, "api.qrserver.com") {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("img")), Header: make(http.Header), Request: req}, nil
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`)), Header: make(http.Header), Request: req}, nil
}

func setupQRTest(t *testing.T, miniApp string) (*handlerpkg.Handler, *bot.Bot, *qrRoundTripper) {
	SetTestEnv(t)
	t.Setenv("MINI_APP_URL", miniApp)
	if err := config.InitConfig(); err != nil {
		t.Fatalf("init config: %v", err)
	}
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}
	link := "https://example.com/sub"
	repo := &StubCustomerRepo{CustomerByTelegramID: &domaincustomer.Customer{TelegramID: 1, SubscriptionLink: &link}}
	h := handlerpkg.NewHandler(nil, nil, tm, repo, nil, nil, nil, nil, nil)
	rt := &qrRoundTripper{}
	http.DefaultTransport = rt
	tgClient := &http.Client{Transport: rt}
	b, err := bot.New("token", bot.WithHTTPClient(time.Second, tgClient), bot.WithSkipGetMe())
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}
	return h, b, rt
}

func TestQRCallbackHandler_NoMiniApp(t *testing.T) {
	h, b, rt := setupQRTest(t, "")
	upd := &models.Update{CallbackQuery: &models.CallbackQuery{From: models.User{ID: 1, LanguageCode: "ru"}, Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}}}}
	h.QRCallbackHandler(context.Background(), b, upd)
	if len(rt.bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(rt.bodies))
	}
	body := rt.bodies[len(rt.bodies)-1]
	if !strings.Contains(body, "Дополнительная") {
		t.Fatalf("caption missing link section: %s", body)
	}
}

func TestQRCallbackHandler_WithMiniApp(t *testing.T) {
	h, b, rt := setupQRTest(t, "https://mini.app")
	upd := &models.Update{CallbackQuery: &models.CallbackQuery{From: models.User{ID: 1, LanguageCode: "ru"}, Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: 1, Chat: models.Chat{ID: 1}}}}}
	h.QRCallbackHandler(context.Background(), b, upd)
	if len(rt.bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(rt.bodies))
	}
	body := rt.bodies[len(rt.bodies)-1]
	if strings.Contains(body, "Дополнительная") {
		t.Fatalf("caption should not contain link section")
	}
	if !strings.Contains(body, "Открыть мини-приложение") || !strings.Contains(body, "https://mini.app") {
		t.Fatalf("mini app button missing: %s", body)
	}
}
