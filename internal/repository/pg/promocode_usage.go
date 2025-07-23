package pg

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type PromocodeUsage struct {
	ID          int64     `db:"id"`
	PromocodeID int64     `db:"promocode_id"`
	UsedBy      int64     `db:"used_by"`
	UsedAt      time.Time `db:"used_at"`
}

type PromocodeUsageRepository struct {
	pool *pgxpool.Pool
}

func NewPromocodeUsageRepository(pool *pgxpool.Pool) *PromocodeUsageRepository {
	return &PromocodeUsageRepository{pool: pool}
}

func (r *PromocodeUsageRepository) Create(ctx context.Context, promoID int64, usedBy int64) error {
	sql, args, err := sq.Insert("promocode_usage").
		Columns("promocode_id", "used_by").
		Values(promoID, usedBy).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert promocode_usage: %w", err)
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *PromocodeUsageRepository) CountByPromocodeID(ctx context.Context, promoID int64) (int, error) {
	sql, args, err := sq.Select("COUNT(*)").From("promocode_usage").Where(sq.Eq{"promocode_id": promoID}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build count promocode_usage: %w", err)
	}
	var count int
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to scan promocode_usage count: %w", err)
	}
	return count, nil
}
