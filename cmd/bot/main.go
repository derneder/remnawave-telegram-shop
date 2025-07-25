package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	crypto "remnawave-tg-shop-bot/internal/adapter/payment/cryptopay"
	"remnawave-tg-shop-bot/internal/adapter/remnawave"
	tgHandler "remnawave-tg-shop-bot/internal/adapter/telegram/handler"
	tgMessenger "remnawave-tg-shop-bot/internal/adapter/telegram/messenger"
	"remnawave-tg-shop-bot/internal/app"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/pkg/translation"
	pg "remnawave-tg-shop-bot/internal/repository/pg"
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

	remClient := remnawave.NewClient(config.RemnawaveUrl(), config.RemnawaveToken(), config.RemnawaveMode())
	cryptoClient := crypto.NewCryptoPayClient(config.CryptoPayUrl(), config.CryptoPayToken())
	messenger := tgMessenger.NewBotMessenger(a.Bot)

	paySvc := payment.NewPaymentService(tm, purchaseRepo, remClient, customerRepo, messenger,
		[]payment.Provider{
			payment.NewCryptoPayProvider(purchaseRepo, cryptoClient),
			payment.NewTributeProvider(purchaseRepo),
		},
		referralRepo, promoRepo, promoUsageRepo, a.Cache)

	syncSvc := syncsvc.NewSyncService(remClient, customerRepo)

	h := tgHandler.NewHandler(syncSvc, paySvc, tm, customerRepo, purchaseRepo, referralRepo, promoRepo, promoUsageRepo, a.Cache)

	a.InitHandlers(h)

	a.Start()

	a.Bot.Start(ctx)
}
