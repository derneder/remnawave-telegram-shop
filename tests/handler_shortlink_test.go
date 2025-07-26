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
)

type recordRoundTripper struct{ req *http.Request }

func (r *recordRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r.req = req
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("http://s.io/ok")), Header: make(http.Header), Request: req}, nil
}

type dummyRoundTripper struct{}

func (dummyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("{}")), Header: make(http.Header), Request: req}, nil
}

func TestShortLinkCallbackHandler_RequestToClck(t *testing.T) {
	rt := &recordRoundTripper{}
	oldTransport := http.DefaultTransport
	http.DefaultTransport = rt
	defer func() { http.DefaultTransport = oldTransport }()

	tgClient := &http.Client{Transport: dummyRoundTripper{}}
	b, err := bot.New("token", bot.WithHTTPClient(time.Second, tgClient), bot.WithSkipGetMe())
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	link := "https://example.com"
	h := handlerpkg.NewHandler(nil, nil, &translation.Manager{}, &StubCustomerRepo{CustomerByTelegramID: &domaincustomer.Customer{TelegramID: 1, SubscriptionLink: &link}}, nil, nil, nil, nil, nil, nil)

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{From: models.User{ID: 1, LanguageCode: "en"}}}

	h.ShortLinkCallbackHandler(context.Background(), b, upd)

	if rt.req == nil {
		t.Fatal("no request recorded")
	}
	if !strings.Contains(rt.req.URL.Host, "clck.ru") {
		t.Fatalf("unexpected shortener host %s", rt.req.URL.Host)
	}
}
