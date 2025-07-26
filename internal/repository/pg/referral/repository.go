package referralpg

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	referralrepo "remnawave-tg-shop-bot/internal/repository/referral"
)

type repository struct {
	pool *pgxpool.Pool
}

// New creates new referral repository backed by PostgreSQL.
func New(pool *pgxpool.Pool) referralrepo.Repository { return &repository{pool: pool} }

func (r *repository) Create(ctx context.Context, referrerID, refereeID int64) error {
	query := sq.Insert("referral").
		Columns("referrer_id", "referee_id", "used_at", "bonus_granted").
		Values(referrerID, refereeID, sq.Expr("NOW()"), false).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build insert referral: %w", err)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec insert referral: %w", err)
	}
	return nil
}

func (r *repository) FindByReferee(ctx context.Context, refereeID int64) (*referralrepo.Model, error) {
	query := sq.Select("id", "referrer_id", "referee_id", "used_at", "bonus_granted").
		From("referral").
		Where(sq.Eq{"referee_id": refereeID}).
		Limit(1).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select referral: %w", err)
	}

	var m referralrepo.Model
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&m.ID, &m.ReferrerID, &m.RefereeID, &m.CreatedAt, &m.BonusGranted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query referral: %w", err)
	}
	return &m, nil
}

func (r *repository) MarkBonusGranted(ctx context.Context, referralID int64) error {
	query := sq.Update("referral").
		Set("bonus_granted", true).
		Where(sq.Eq{"id": referralID}).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build update referral: %w", err)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec update referral: %w", err)
	}
	return nil
}

func (r *repository) CountByReferrer(ctx context.Context, referrerID int64) (int, error) {
	query := sq.Select("count(*)").
		From("referral").
		Where(sq.Eq{"referrer_id": referrerID}).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build count referral: %w", err)
	}

	var cnt int
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&cnt); err != nil {
		return 0, fmt.Errorf("exec count referral: %w", err)
	}
	return cnt, nil
}
