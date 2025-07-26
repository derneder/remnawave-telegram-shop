package pg

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Promocode struct {
	ID        int64     `db:"id"`
	Code      string    `db:"code"`
	Months    int       `db:"months"`
	Type      int16     `db:"type"`
	Days      int       `db:"days"`
	Amount    int       `db:"amount"`
	UsesLeft  int       `db:"uses_left"`
	CreatedBy int64     `db:"created_by"`
	CreatedAt time.Time `db:"created_at"`
	Active    bool      `db:"active"`
	Deleted   bool      `db:"deleted"`
}

type PromocodeRepository struct {
	pool *pgxpool.Pool
}

func NewPromocodeRepository(pool *pgxpool.Pool) *PromocodeRepository {
	return &PromocodeRepository{pool: pool}
}

func (r *PromocodeRepository) Create(ctx context.Context, promo *Promocode) (*Promocode, error) {
	sql, args, err := sq.Insert("promocode").
		Columns("code", "months", "uses_left", "created_by", "active", "type", "days", "amount").
		Values(promo.Code, promo.Months, promo.UsesLeft, promo.CreatedBy, promo.Active, promo.Type, promo.Days, promo.Amount).
		Suffix("RETURNING id, created_at, active").
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build insert promocode: %w", err)
	}
	row := r.pool.QueryRow(ctx, sql, args...)
	if err := row.Scan(&promo.ID, &promo.CreatedAt, &promo.Active); err != nil {
		return nil, err
	}
	return promo, nil
}

func (r *PromocodeRepository) GetByCode(ctx context.Context, code string) (*Promocode, error) {
	sql, args, err := sq.Select("id", "code", "months", "type", "days", "amount", "uses_left", "created_by", "created_at", "active").
		From("promocode").
		Where(sq.Eq{"code": code, "deleted": false}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select promocode by code: %w", err)
	}
	promo := &Promocode{}
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&promo.ID, &promo.Code, &promo.Months, &promo.Type, &promo.Days, &promo.Amount, &promo.UsesLeft, &promo.CreatedBy, &promo.CreatedAt, &promo.Active)
	if err != nil {
		return nil, err
	}
	return promo, nil
}

func (r *PromocodeRepository) GetById(ctx context.Context, id int64) (*Promocode, error) {
	sql, args, err := sq.Select("id", "code", "months", "type", "days", "amount", "uses_left", "created_by", "created_at", "active").
		From("promocode").
		Where(sq.Eq{"id": id, "deleted": false}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select promocode by id: %w", err)
	}
	promo := &Promocode{}
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&promo.ID, &promo.Code, &promo.Months, &promo.Type, &promo.Days, &promo.Amount, &promo.UsesLeft, &promo.CreatedBy, &promo.CreatedAt, &promo.Active)
	if err != nil {
		return nil, err
	}
	return promo, nil
}

func (r *PromocodeRepository) DecrementUses(ctx context.Context, id int64) error {
	sql, args, err := sq.Update("promocode").
		Set("uses_left", sq.Expr("uses_left - 1")).
		Where(sq.Eq{"id": id, "deleted": false}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update promocode: %w", err)
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *PromocodeRepository) UpdateStatus(ctx context.Context, id int64, active bool) error {
	sql, args, err := sq.Update("promocode").
		Set("active", active).
		Where(sq.Eq{"id": id, "deleted": false}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update promocode status: %w", err)
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *PromocodeRepository) UpdateDeleteStatus(ctx context.Context, id int64, deleted bool) error {
	sql, args, err := sq.Update("promocode").
		Set("deleted", deleted).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update promocode deleted: %w", err)
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *PromocodeRepository) FindByCreator(ctx context.Context, createdBy int64) ([]Promocode, error) {
	sql, args, err := sq.Select("id", "code", "months", "type", "days", "amount", "uses_left", "created_by", "created_at", "active").
		From("promocode").
		Where(sq.Eq{"created_by": createdBy, "deleted": false}).
		Where(sq.Or{sq.Gt{"uses_left": 0}, sq.Eq{"uses_left": 0}}).
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select promocodes: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query promocodes: %w", err)
	}
	defer rows.Close()

	var list []Promocode
	for rows.Next() {
		var p Promocode
		if err := rows.Scan(&p.ID, &p.Code, &p.Months, &p.Type, &p.Days, &p.Amount, &p.UsesLeft, &p.CreatedBy, &p.CreatedAt, &p.Active); err != nil {
			return nil, fmt.Errorf("failed to scan promocode: %w", err)
		}
		list = append(list, p)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating promocode rows: %w", rows.Err())
	}
	return list, nil
}
