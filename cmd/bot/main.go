package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"remnawave-tg-shop-bot/internal/app"
)

// Version is set via -ldflags.
var Version = "dev"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	slog.Info("starting bot", "version", Version)

	a, err := app.New(ctx)
	if err != nil {
		slog.Error("init app", "err", err)
		return
	}
	defer a.Shutdown(ctx)

	a.Start()

	a.Bot.Start(ctx)
}
