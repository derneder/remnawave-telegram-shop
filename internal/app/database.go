package app

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

// InitDatabase creates a connection pool with sane defaults.
func InitDatabase(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 20
	cfg.MinConns = 5
	return pgxpool.ConnectConfig(ctx, cfg)
}
