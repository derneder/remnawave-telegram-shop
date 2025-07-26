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
	pg "remnawave-tg-shop-bot/internal/repository/pg"
	"remnawave-tg-shop-bot/internal/service/notification"
)

const cronStopTimeout = 5 * time.Second

// App groups dependencies of the bot.
type App struct {
	Bot   *bot.Bot
	Pool  *pgxpool.Pool
	Cron  *cron.Cron
	Cache *cache.Cache
}

func New(ctx context.Context) (*App, error) {

	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		return nil, fmt.Errorf("init translations: %w", err)
	}

	pool, err := InitDatabase(ctx, config.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}

	if err := pg.RunMigrations(ctx, &pg.MigrationConfig{
		Direction:      "up",
		MigrationsPath: "./db/migrations",
	}, pool); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
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
	if u, err := b.GetMe(ctx); err == nil {
		config.SetBotURL("https://t.me/" + u.Username)
	}
	customerRepo := pg.NewCustomerRepository(pool)
	subSvc := notification.NewSubscriptionService(customerRepo, b, tm)

	sched := cron.New(cron.WithLocation(time.UTC))
	if err := notification.RegisterSubscriptionCron(sched, subSvc); err != nil {
		return nil, fmt.Errorf("schedule subscription cron: %w", err)
	}

	metricsSrv := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.GetHealthCheckPort()),
		Handler:           observability.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
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
	cache := cache.NewCache(ctx, time.Hour)

	return &App{Bot: b, Pool: pool, Cron: sched, Cache: cache}, nil
}

func (a *App) Start() {
	if a.Cron != nil {
		a.Cron.Start()
	}
}

func (a *App) Shutdown(ctx context.Context) {
	if a.Cache != nil {
		defer a.Cache.Close()
	}

	if a.Cron != nil {
		stopCtx := a.Cron.Stop()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), cronStopTimeout)
		defer cancel()
		select {
		case <-stopCtx.Done():
		case <-timeoutCtx.Done():
			slog.Warn("cron stop timeout")
		case <-ctx.Done():
		}
	}
	if a.Bot != nil {
		_, _ = a.Bot.Close(ctx)
	}
	a.Pool.Close()
}
