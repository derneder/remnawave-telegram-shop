package messenger

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Messenger abstracts Telegram bot actions used by services.
type Messenger interface {
	SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
	DeleteMessage(ctx context.Context, params *bot.DeleteMessageParams) (bool, error)
	CreateInvoiceLink(ctx context.Context, params *bot.CreateInvoiceLinkParams) (string, error)
}

type BotMessenger struct {
	b *bot.Bot
}

func NewBotMessenger(b *bot.Bot) *BotMessenger { return &BotMessenger{b: b} }

func (m *BotMessenger) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	return m.b.SendMessage(ctx, params)
}

func (m *BotMessenger) DeleteMessage(ctx context.Context, params *bot.DeleteMessageParams) (bool, error) {
	return m.b.DeleteMessage(ctx, params)
}

func (m *BotMessenger) CreateInvoiceLink(ctx context.Context, params *bot.CreateInvoiceLinkParams) (string, error) {
	return m.b.CreateInvoiceLink(ctx, params)
}
