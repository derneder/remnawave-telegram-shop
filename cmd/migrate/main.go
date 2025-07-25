package main

import (
	"context"
	"log"

	"remnawave-tg-shop-bot/internal/app"
	"remnawave-tg-shop-bot/internal/pkg/config"
	"remnawave-tg-shop-bot/internal/repository/pg"
)

func main() {
	if err := config.InitConfig(); err != nil {
		log.Fatalf("init config: %v", err)
	}
	ctx := context.Background()

	pool, err := app.InitDatabase(ctx, config.DatabaseURL())
	if err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer pool.Close()

	if err := pg.RunMigrations(ctx, &pg.MigrationConfig{
		Direction:      "up",
		MigrationsPath: "./db/migrations",
		Steps:          0,
	}, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}
