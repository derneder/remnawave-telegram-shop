package handler

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
)

func (h *Handler) SyncUsersCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if err := h.syncService.Sync(ctx); err != nil {
		slog.Error("Error syncing users", "err", err)
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Users synced",
	})
	if err != nil {
		slog.Error("Error sending sync message", "err", err)
	}
}
