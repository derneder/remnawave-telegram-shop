package promotion

import (
	"context"

	"remnawave-tg-shop-bot/internal/repository/pg"
)

// Repository provides persistence for promocodes.
type Repository interface {
	Create(ctx context.Context, promo *pg.Promocode) (*pg.Promocode, error)
	UpdateStatus(ctx context.Context, id int64, active bool) error
	UpdateDeleteStatus(ctx context.Context, id int64, deleted bool) error
}
