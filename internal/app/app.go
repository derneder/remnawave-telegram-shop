package app

import (
	"context"
	"net/http"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/robfig/cron/v3"

	"remnawave-tg-shop-bot/internal/observability"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

// App groups dependencies of the bot.
type App struct {
	Bot  *bot.Bot
	Pool *pgxpool.Pool
	Cron *cron.Cron
}

func New(ctx context.Context) (*App, error) {
	config.InitConfig()

	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		return nil, err
	}

	pool, err := InitDatabase(ctx, config.DatabaseURL())
	if err != nil {
		return nil, err
	}

	b, err := bot.New(config.TelegramToken(), bot.WithMiddlewares(
		func(next bot.HandlerFunc) bot.HandlerFunc {
			return func(ctx context.Context, b *bot.Bot, update *models.Update) {
				start := time.Now()
				next(ctx, b, update)
				observability.RequestDuration.WithLabelValues("telegram").Observe(time.Since(start).Seconds())
			}
		},
	))
	if err != nil {
		return nil, err
	}

	go func() {
		http.ListenAndServe(":9100", observability.Handler())
	}()

	return &App{Bot: b, Pool: pool, Cron: cron.New()}, nil
}

func (a *App) Close() {
	a.Pool.Close()
}
