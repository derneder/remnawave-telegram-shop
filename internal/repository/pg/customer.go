package pg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"remnawave-tg-shop-bot/utils"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	domain "remnawave-tg-shop-bot/internal/domain/customer"
)

type CustomerRepository struct {
	pool *pgxpool.Pool
}

func NewCustomerRepository(pool *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{pool: pool}
}

type Customer = domain.Customer

func (cr *CustomerRepository) FindByExpirationRange(ctx context.Context, startDate, endDate time.Time) (*[]Customer, error) {
	buildSelect := sq.Select("id", "telegram_id", "expire_at", "created_at", "subscription_link", "language", "balance").
		From("customer").
		Where(
			sq.And{
				sq.NotEq{"expire_at": nil},
				sq.GtOrEq{"expire_at": startDate},
				sq.LtOrEq{"expire_at": endDate},
			},
		).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := buildSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := cr.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query customers by expiration range: %w", err)
	}
	defer rows.Close()

	var customers []Customer
	for rows.Next() {
		var customer Customer
		err := rows.Scan(
			&customer.ID,
			&customer.TelegramID,
			&customer.ExpireAt,
			&customer.CreatedAt,
			&customer.SubscriptionLink,
			&customer.Language,
			&customer.Balance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer row: %w", err)
		}
		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over customer rows: %w", err)
	}

	return &customers, nil
}

func (cr *CustomerRepository) FindById(ctx context.Context, id int64) (*Customer, error) {
	buildSelect := sq.Select("id", "telegram_id", "expire_at", "created_at", "subscription_link", "language", "balance").
		From("customer").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := buildSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var customer Customer

	err = cr.pool.QueryRow(ctx, sql, args...).Scan(
		&customer.ID,
		&customer.TelegramID,
		&customer.ExpireAt,
		&customer.CreatedAt,
		&customer.SubscriptionLink,
		&customer.Language,
		&customer.Balance,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query customer: %w", err)
	}
	return &customer, nil
}

func (cr *CustomerRepository) FindByTelegramId(ctx context.Context, telegramId int64) (*Customer, error) {
	buildSelect := sq.Select("id", "telegram_id", "expire_at", "created_at", "subscription_link", "language", "balance").
		From("customer").
		Where(sq.Eq{"telegram_id": telegramId}).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := buildSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var customer Customer

	err = cr.pool.QueryRow(ctx, sql, args...).Scan(
		&customer.ID,
		&customer.TelegramID,
		&customer.ExpireAt,
		&customer.CreatedAt,
		&customer.SubscriptionLink,
		&customer.Language,
		&customer.Balance,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query customer: %w", err)
	}
	return &customer, nil
}

func (cr *CustomerRepository) Create(ctx context.Context, customer *Customer) (*Customer, error) {
	buildInsert := sq.Insert("customer").
		Columns("telegram_id", "expire_at", "language", "balance").
		PlaceholderFormat(sq.Dollar).
		Values(customer.TelegramID, customer.ExpireAt, customer.Language, customer.Balance).
		Suffix("RETURNING id, created_at")
	sqlStr, args, err := buildInsert.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build insert query: %w", err)
	}

	row := cr.pool.QueryRow(ctx, sqlStr, args...)
	var id int64
	var createdAt time.Time
	if err := row.Scan(&id, &createdAt); err != nil {
		return nil, fmt.Errorf("failed to insert customer: %w", err)
	}
	customer.ID = id
	customer.CreatedAt = createdAt

	slog.Info("user created in bot database", "telegramId", utils.MaskHalfInt64(customer.TelegramID))
	return customer, nil
}

func (cr *CustomerRepository) UpdateFields(ctx context.Context, id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	buildUpdate := sq.Update("customer").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": id})

	for field, value := range updates {
		buildUpdate = buildUpdate.Set(field, value)
	}

	sql, args, err := buildUpdate.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	tx, err := cr.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no customer found with id: %s", utils.MaskHalfInt64(id))
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (cr *CustomerRepository) FindByTelegramIds(ctx context.Context, telegramIDs []int64) ([]Customer, error) {
	buildSelect := sq.Select("id", "telegram_id", "expire_at", "created_at", "subscription_link", "language", "balance").
		From("customer").
		Where(sq.Eq{"telegram_id": telegramIDs}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := buildSelect.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := cr.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query customers: %w", err)
	}
	defer rows.Close()

	var customers []Customer
	for rows.Next() {
		var customer Customer
		err := rows.Scan(
			&customer.ID,
			&customer.TelegramID,
			&customer.ExpireAt,
			&customer.CreatedAt,
			&customer.SubscriptionLink,
			&customer.Language,
			&customer.Balance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer row: %w", err)
		}
		customers = append(customers, customer)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over customer rows: %w", err)
	}

	return customers, nil
}

func (cr *CustomerRepository) CreateBatch(ctx context.Context, customers []Customer) error {
	if len(customers) == 0 {
		return nil
	}
	builder := sq.Insert("customer").
		Columns("telegram_id", "expire_at", "language", "subscription_link", "balance").
		PlaceholderFormat(sq.Dollar)
	for _, cust := range customers {
		builder = builder.Values(cust.TelegramID, cust.ExpireAt, cust.Language, cust.SubscriptionLink, cust.Balance)
	}
	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build batch insert query: %w", err)
	}

	tx, err := cr.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (cr *CustomerRepository) UpdateBatch(ctx context.Context, customers []Customer) error {
	if len(customers) == 0 {
		return nil
	}
	query := "UPDATE customer SET expire_at = c.expire_at, language = c.language, subscription_link = c.subscription_link, balance = c.balance FROM (VALUES "
	var args []interface{}
	for i, cust := range customers {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("($%d::bigint, $%d::timestamp, $%d::text, $%d::text, $%d::numeric)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5)
		args = append(args, cust.TelegramID, cust.ExpireAt, cust.Language, cust.SubscriptionLink, cust.Balance)
	}
	query += ") AS c(telegram_id, expire_at, language, subscription_link, balance) WHERE customer.telegram_id = c.telegram_id"

	tx, err := cr.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute batch update: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (cr *CustomerRepository) DeleteByNotInTelegramIds(ctx context.Context, telegramIDs []int64) error {
	var buildDeleteUsers sq.DeleteBuilder
	var buildDeletePromo sq.DeleteBuilder
	var buildDeletePromoUsage sq.DeleteBuilder

	if len(telegramIDs) == 0 {
		buildDeleteUsers = sq.Delete("customer")
		buildDeletePromo = sq.Delete("promocode")
		buildDeletePromoUsage = sq.Delete("promocode_usage")
	} else {
		normalPromo := sq.Select("id").
			From("promocode").
			PlaceholderFormat(sq.Dollar).
			Where(sq.Eq{"created_by": telegramIDs})

		sqlNormalPromoStr, args, err := normalPromo.ToSql()
		if err != nil {
			return fmt.Errorf("failed to build select query: %w", err)
		}

		promoIDs := make([]int, 0)
		promoRows, err := cr.pool.Query(ctx, sqlNormalPromoStr, args...)
		if err != nil {
			return fmt.Errorf("failed to collect promo id from select query: %w", err)
		}
		for promoRows.Next() {
			var id int
			promoRows.Scan(&id)
			promoIDs = append(promoIDs, id)
		}

		buildDeleteUsers = sq.Delete("customer").
			PlaceholderFormat(sq.Dollar).
			Where(sq.NotEq{"telegram_id": telegramIDs})
		buildDeletePromo = sq.Delete("promocode").
			PlaceholderFormat(sq.Dollar).
			Where(sq.NotEq{"created_by": telegramIDs})

		if len(promoIDs) == 0 {
			buildDeletePromoUsage = sq.Delete("promocode_usage")
		} else {
			buildDeletePromoUsage = sq.Delete("promocode_usage").
				PlaceholderFormat(sq.Dollar).
				Where(sq.NotEq{"id": promoIDs})
		}

	}

	sqlStr, args, err := buildDeletePromoUsage.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete promo_usage query: %w", err)
	}

	_, err = cr.pool.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("failed to delete promo_usage: %w", err)
	}

	sqlStr, args, err = buildDeletePromo.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete promo query: %w", err)
	}

	_, err = cr.pool.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("failed to delete promo: %w", err)
	}

	sqlStr, args, err = buildDeleteUsers.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete customer query: %w", err)
	}

	_, err = cr.pool.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	return nil
}
