package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	crypto "remnawave-tg-shop-bot/internal/adapter/payment/cryptopay"
	tributewebhook "remnawave-tg-shop-bot/internal/adapter/payment/tribute"
	"remnawave-tg-shop-bot/internal/adapter/remnawave"
	tgHandler "remnawave-tg-shop-bot/internal/adapter/telegram/handler"
	tgMessenger "remnawave-tg-shop-bot/internal/adapter/telegram/messenger"
	"remnawave-tg-shop-bot/internal/app"
	"remnawave-tg-shop-bot/internal/observability"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	pg "remnawave-tg-shop-bot/internal/repository/pg"
	custsvc "remnawave-tg-shop-bot/internal/service/customer"
	"remnawave-tg-shop-bot/internal/service/payment"
	syncsvc "remnawave-tg-shop-bot/internal/service/sync"
)

// Version is set via -ldflags.
var Version = "dev"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := config.InitConfig(); err != nil {
		slog.Error("init config", "err", err)
		return
	}

	slog.Info("starting bot", "version", Version)

	a, err := app.New(ctx)
	if err != nil {
		slog.Error("init app", "err", err)
		return
	}
	defer a.Shutdown(ctx)

	tm := translation.GetInstance()
	customerRepo := pg.NewCustomerRepository(a.Pool)
	purchaseRepo := pg.NewPurchaseRepository(a.Pool)
	referralRepo := pg.NewReferralRepository(a.Pool)
	promoRepo := pg.NewPromocodeRepository(a.Pool)
	promoUsageRepo := pg.NewPromocodeUsageRepository(a.Pool)

	remClient, err := remnawave.NewClient(config.RemnawaveUrl(), config.RemnawaveToken(), config.RemnawaveMode())
	if err != nil {
		slog.Error("init remnawave client", "err", err)
		return
	}
	cryptoClient := crypto.NewCryptoPayClient(config.CryptoPayUrl(), config.CryptoPayToken())
	messenger := tgMessenger.NewBotMessenger(a.Bot)

	svcCustomer := custsvc.NewService(customerRepo)

	paySvc := payment.NewPaymentService(tm, purchaseRepo, remClient, customerRepo, messenger,
		[]payment.Provider{
			payment.NewCryptoPayProvider(purchaseRepo, cryptoClient),
			payment.NewTributeProvider(purchaseRepo),
		},
		referralRepo, promoRepo, promoUsageRepo, a.Cache)

	syncSvc := syncsvc.NewSyncService(remClient, customerRepo)

	h := tgHandler.NewHandler(syncSvc, paySvc, tm, customerRepo, purchaseRepo, referralRepo, promoRepo, promoUsageRepo, a.Cache)

	httpMux := http.NewServeMux()
	httpMux.Handle("/healthcheck", observability.Handler())
	cfg := config.Tribute()
	if cfg.TributeWebhookPath != "" {
		httpMux.Handle(cfg.TributeWebhookPath, tributewebhook.NewHandler(cfg.TributeAPIKey, svcCustomer))
	}

	go func() {
		srv := &http.Server{Addr: fmt.Sprintf(":%d", config.GetHealthCheckPort()), Handler: httpMux, ReadHeaderTimeout: 5 * time.Second}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server", "err", err)
		}
	}()

	h.Start(ctx)

	a.InitHandlers(h)

	a.Start()

	a.Bot.Start(ctx)
}
