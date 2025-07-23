package main

import (
	"context"
	"os"
	"os/signal"

	"remnawave-tg-shop-bot/internal/app"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	a, err := app.New(ctx)
	if err != nil {
		panic(err)
	}
	defer a.Close()

	a.Bot.Start(ctx)
}
