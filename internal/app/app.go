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

	"remnawave-tg-shop-bot/internal/adapter/payment/cryptopay"
	"remnawave-tg-shop-bot/internal/adapter/payment/yookassa"
	"remnawave-tg-shop-bot/internal/adapter/remnawave"
	tgHandler "remnawave-tg-shop-bot/internal/adapter/telegram/handler"
	tgMessenger "remnawave-tg-shop-bot/internal/adapter/telegram/messenger"
	"remnawave-tg-shop-bot/internal/observability"
	"remnawave-tg-shop-bot/internal/pkg/cache"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	pgrepo "remnawave-tg-shop-bot/internal/repository/pg"
	"remnawave-tg-shop-bot/internal/service/payment"
	syncsvc "remnawave-tg-shop-bot/internal/service/sync"
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

	customerRepo := pgrepo.NewCustomerRepository(pool)
	purchaseRepo := pgrepo.NewPurchaseRepository(pool)
	referralRepo := pgrepo.NewReferralRepository(pool)
	promoRepo := pgrepo.NewPromocodeRepository(pool)
	promoUsageRepo := pgrepo.NewPromocodeUsageRepository(pool)
	c := cache.NewCache(time.Minute)

	remClient := remnawave.NewClient(config.RemnawaveUrl(), config.RemnawaveToken(), config.RemnawaveMode())
	messenger := tgMessenger.NewBotMessenger(b)
	cryptoClient := cryptopay.NewCryptoPayClient(config.CryptoPayUrl(), config.CryptoPayToken())
	yookasaClient := yookasa.NewClient(config.YookasaUrl(), config.YookasaShopId(), config.YookasaSecretKey())

	paymentSvc := payment.NewPaymentService(tm, purchaseRepo, remClient, customerRepo, messenger, cryptoClient, yookasaClient, referralRepo, promoRepo, promoUsageRepo, c)
	syncSvc := syncsvc.NewSyncService(remClient, customerRepo)

	h := tgHandler.NewHandler(syncSvc, paymentSvc, tm, customerRepo, purchaseRepo, cryptoClient, yookasaClient, referralRepo, promoRepo, promoUsageRepo, c)
	initHandlers(b, h)

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
	if a.Cron != nil {
		a.Cron.Stop()
	}
	a.Pool.Close()
	if a.Cache != nil {
		a.Cache.Close()
	}
}
