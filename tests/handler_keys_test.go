package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	handlerpkg "remnawave-tg-shop-bot/internal/adapter/telegram/handler"
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

type countRoundTripper struct{ calls int }

func (c *countRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	c.calls++
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Header: make(http.Header), Request: req}, nil
}

func TestKeysCallbackHandler_InvalidScheme(t *testing.T) {
	SetTestEnv(t)
	rt := &countRoundTripper{}
	oldTransport := http.DefaultTransport
	http.DefaultTransport = rt
	defer func() { http.DefaultTransport = oldTransport }()

	tgClient := &http.Client{Transport: rt}
	b, err := bot.New("token", bot.WithHTTPClient(time.Second, tgClient), bot.WithSkipGetMe())
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	link := "http://example.com/file.txt"
	h := handlerpkg.NewHandler(nil, nil, &translation.Manager{}, &StubCustomerRepo{CustomerByTelegramID: &domaincustomer.Customer{TelegramID: 1, SubscriptionLink: &link}}, nil, nil, nil, nil, nil)

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{From: models.User{ID: 1, LanguageCode: "en"}}}

	h.KeysCallbackHandler(context.Background(), b, upd)

	if rt.calls != 0 {
		t.Fatalf("unexpected http calls: %d", rt.calls)
	}
}

func TestKeysCallbackHandler_HostNotAllowed(t *testing.T) {
	SetTestEnv(t)
	rt := &countRoundTripper{}
	oldTransport := http.DefaultTransport
	http.DefaultTransport = rt
	defer func() { http.DefaultTransport = oldTransport }()

	tgClient := &http.Client{Transport: rt}
	b, err := bot.New("token", bot.WithHTTPClient(time.Second, tgClient), bot.WithSkipGetMe())
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	link := "https://evil.com/file.txt"
	h := handlerpkg.NewHandler(nil, nil, &translation.Manager{}, &StubCustomerRepo{CustomerByTelegramID: &domaincustomer.Customer{TelegramID: 1, SubscriptionLink: &link}}, nil, nil, nil, nil, nil)

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{From: models.User{ID: 1, LanguageCode: "en"}}}

	h.KeysCallbackHandler(context.Background(), b, upd)

	if rt.calls != 0 {
		t.Fatalf("unexpected http calls: %d", rt.calls)
	}
}
