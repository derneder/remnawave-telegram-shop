package handler_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	handlerpkg "remnawave-tg-shop-bot/internal/adapter/telegram/handler"
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	"remnawave-tg-shop-bot/tests/testutils"
)

type closeRecorder struct {
	closed *bool
}

func (c *closeRecorder) Read(p []byte) (int, error) { return 0, io.EOF }
func (c *closeRecorder) Close() error               { *c.closed = true; return nil }

type testRoundTripper struct {
	mu                 sync.Mutex
	call               int
	firstClosed        bool
	closedBeforeSecond bool
}

func (t *testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.call++
	switch t.call {
	case 1:
		body := &closeRecorder{closed: &t.firstClosed}
		return &http.Response{StatusCode: http.StatusInternalServerError, Body: body, Header: make(http.Header), Request: req}, nil
	case 2:
		t.closedBeforeSecond = t.firstClosed
		body := io.NopCloser(strings.NewReader("http://s.io/ok"))
		return &http.Response{StatusCode: http.StatusOK, Body: body, Header: make(http.Header), Request: req}, nil
	default:
		return nil, io.EOF
	}
}

type dummyRoundTripper struct{}

func (dummyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("{}")), Header: make(http.Header), Request: req}, nil
}

func TestShortLinkCallbackHandler_BodyClosedOnRetry(t *testing.T) {
	rt := &testRoundTripper{}
	oldTransport := http.DefaultTransport
	http.DefaultTransport = rt
	defer func() { http.DefaultTransport = oldTransport }()

	tgClient := &http.Client{Transport: dummyRoundTripper{}}
	b, err := bot.New("token", bot.WithHTTPClient(time.Second, tgClient), bot.WithSkipGetMe())
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	link := "https://example.com"
	h := handlerpkg.NewHandler(nil, nil, &translation.Manager{}, &testutils.StubCustomerRepo{CustomerByTelegramID: &domaincustomer.Customer{TelegramID: 1, SubscriptionLink: &link}}, nil, nil, nil, nil, nil)

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{From: models.User{ID: 1, LanguageCode: "en"}}}

	h.ShortLinkCallbackHandler(context.Background(), b, upd)

	if !rt.firstClosed {
		t.Fatal("first response body not closed")
	}
	if !rt.closedBeforeSecond {
		t.Fatal("response body closed after retry request")
	}
}
