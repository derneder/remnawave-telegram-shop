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
)

type safeRecordClient struct {
	calls  int
	status int
	resp   string
}

func (c *safeRecordClient) Do(req *http.Request) (*http.Response, error) {
	c.calls++
	body := io.NopCloser(strings.NewReader(c.resp))
	return &http.Response{StatusCode: c.status, Body: body, Header: make(http.Header), Request: req}, nil
}

func newBotSafe(c *safeRecordClient) (*bot.Bot, error) {
	return bot.New("token", bot.WithHTTPClient(time.Second, c), bot.WithSkipGetMe())
}

func TestSafeEditMessageText_Skip(t *testing.T) {
	client := &safeRecordClient{status: http.StatusOK, resp: `{"ok":true,"result":{"message_id":1}}`}
	b, err := newBotSafe(client)
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	kb := [][]models.InlineKeyboardButton{{{Text: "a", CallbackData: "1"}}}
	old := &models.Message{Text: "text", ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: kb}}
	params := &bot.EditMessageTextParams{ChatID: 1, MessageID: 1, Text: "text", ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}}

	if _, err := handlerpkg.SafeEditMessageText(context.Background(), b, old, params); err != nil {
		t.Fatalf("SafeEditMessageText returned error: %v", err)
	}
	if client.calls != 0 {
		t.Fatalf("expected 0 calls, got %d", client.calls)
	}
}

func TestSafeEditMessageText_Call(t *testing.T) {
	client := &safeRecordClient{status: http.StatusOK, resp: `{"ok":true,"result":{"message_id":1}}`}
	b, err := newBotSafe(client)
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	kb := [][]models.InlineKeyboardButton{{{Text: "a", CallbackData: "1"}}}
	old := &models.Message{Text: "text", ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: kb}}
	params := &bot.EditMessageTextParams{ChatID: 1, MessageID: 1, Text: "new", ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: kb}}

	if _, err := handlerpkg.SafeEditMessageText(context.Background(), b, old, params); err != nil {
		t.Fatalf("SafeEditMessageText returned error: %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("expected 1 call, got %d", client.calls)
	}
}

func TestSafeEditMessageText_IgnoreError(t *testing.T) {
	resp := `{"ok":false,"error_code":400,"description":"Bad Request: message is not modified"}`
	client := &safeRecordClient{status: http.StatusOK, resp: resp}
	b, err := newBotSafe(client)
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	params := &bot.EditMessageTextParams{ChatID: 1, MessageID: 1, Text: "t"}
	if _, err := handlerpkg.SafeEditMessageText(context.Background(), b, nil, params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("expected 1 call, got %d", client.calls)
	}
}
