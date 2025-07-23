package handler

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
	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/pkg/translation"
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

type stubCustomerRepoSL struct{}

func (stubCustomerRepoSL) FindById(context.Context, int64) (*domaincustomer.Customer, error) {
	return nil, nil
}
func (stubCustomerRepoSL) FindByTelegramId(context.Context, int64) (*domaincustomer.Customer, error) {
	link := "https://example.com"
	return &domaincustomer.Customer{TelegramID: 1, SubscriptionLink: &link}, nil
}
func (stubCustomerRepoSL) Create(ctx context.Context, c *domaincustomer.Customer) (*domaincustomer.Customer, error) {
	return c, nil
}
func (stubCustomerRepoSL) UpdateFields(context.Context, int64, map[string]interface{}) error {
	return nil
}
func (stubCustomerRepoSL) FindByTelegramIds(context.Context, []int64) ([]domaincustomer.Customer, error) {
	return nil, nil
}
func (stubCustomerRepoSL) DeleteByNotInTelegramIds(context.Context, []int64) error      { return nil }
func (stubCustomerRepoSL) CreateBatch(context.Context, []domaincustomer.Customer) error { return nil }
func (stubCustomerRepoSL) UpdateBatch(context.Context, []domaincustomer.Customer) error { return nil }
func (stubCustomerRepoSL) FindByExpirationRange(context.Context, time.Time, time.Time) (*[]domaincustomer.Customer, error) {
	return nil, nil
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

	h := &Handler{
		customerRepository: stubCustomerRepoSL{},
		shortLinks:         make(map[int64][]ShortLink),
		translation:        &translation.Manager{},
	}

	upd := &models.Update{CallbackQuery: &models.CallbackQuery{From: models.User{ID: 1, LanguageCode: "en"}}}

	h.ShortLinkCallbackHandler(context.Background(), b, upd)

	if !rt.firstClosed {
		t.Fatal("first response body not closed")
	}
	if !rt.closedBeforeSecond {
		t.Fatal("response body closed after retry request")
	}
}
