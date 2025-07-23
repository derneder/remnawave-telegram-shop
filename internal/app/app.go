package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/robfig/cron/v3"
	"log/slog"

	"remnawave-tg-shop-bot/internal/observability"
	"remnawave-tg-shop-bot/internal/pkg/cache"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

// App groups dependencies of the bot.
type App struct {
	Bot   *bot.Bot
	Pool  *pgxpool.Pool
	Cron  *cron.Cron
	Cache *cache.Cache
}

func New(ctx context.Context) (*App, error) {
	config.InitConfig()

	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		return nil, fmt.Errorf("init translations: %w", err)
	}

	pool, err := InitDatabase(ctx, config.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
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
		return nil, fmt.Errorf("create bot: %w", err)
	}

	metricsSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.GetHealthCheckPort()),
		Handler: observability.Handler(),
	}

	go func() {
		if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("metrics server", "err", err)
		}
	}()

	go func() {
		<-ctx.Done()
		_ = metricsSrv.Shutdown(context.Background())
	}()

	c := cache.NewCache(time.Hour)

	return &App{Bot: b, Pool: pool, Cron: cron.New(), Cache: c}, nil
}

func (a *App) Close() {
	a.Pool.Close()
	if a.Cache != nil {
		a.Cache.Close()
	}
}
