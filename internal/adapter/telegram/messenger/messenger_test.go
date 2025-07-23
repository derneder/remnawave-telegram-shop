package messenger_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnawave-tg-shop-bot/internal/adapter/telegram/messenger"
)

type recordClient struct{ calls int }

func (c *recordClient) Do(req *http.Request) (*http.Response, error) {
	c.calls++
	var resp string
	switch {
	case strings.Contains(req.URL.Path, "sendMessage"):
		resp = `{"ok":true,"result":{"message_id":1}}`
	case strings.Contains(req.URL.Path, "deleteMessage"):
		resp = `{"ok":true,"result":true}`
	default:
		resp = `{"ok":true,"result":"link"}`
	}
	body := io.NopCloser(strings.NewReader(resp))
	return &http.Response{StatusCode: http.StatusOK, Body: body, Header: make(http.Header), Request: req}, nil
}

func newBot(c *recordClient) (*bot.Bot, error) {
	return bot.New("token", bot.WithHTTPClient(time.Second, c), bot.WithSkipGetMe())
}

func TestMessengerMethods(t *testing.T) {
	rc := &recordClient{}
	b, err := newBot(rc)
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}
	m := messenger.NewBotMessenger(b)

	if _, err = m.SendMessage(context.Background(), &bot.SendMessageParams{ChatID: 1, Text: "hi"}); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if _, err = m.DeleteMessage(context.Background(), &bot.DeleteMessageParams{ChatID: 1, MessageID: 1}); err != nil {
		t.Fatalf("DeleteMessage: %v", err)
	}
	if _, err = m.CreateInvoiceLink(context.Background(), &bot.CreateInvoiceLinkParams{Title: "t", Payload: "p", Prices: []models.LabeledPrice{{Label: "l", Amount: 1}}}); err != nil {
		t.Fatalf("CreateInvoiceLink: %v", err)
	}
	if rc.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", rc.calls)
	}
}
