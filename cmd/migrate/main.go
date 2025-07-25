package main

import (
	"context"
	"log/slog"

	"remnawave-tg-shop-bot/internal/app"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/repository/pg"
)

func main() {
	if err := config.InitConfig(); err != nil {
		slog.Error("init config", "err", err)
		return
	}
	ctx := context.Background()

	pool, err := app.InitDatabase(ctx, config.DatabaseURL())
	if err != nil {
		slog.Error("init db", "err", err)
		return
	}
	defer pool.Close()

	if err := pg.RunMigrations(ctx, &pg.MigrationConfig{
		Direction:      "up",
		MigrationsPath: "./db/migrations",
		Steps:          0,
	}, pool); err != nil {
		slog.Error("migrate", "err", err)
		return
	}

	slog.Info("migrations applied successfully")
}
