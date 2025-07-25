package promotion

import (
	"context"

	"remnawave-tg-shop-bot/internal/repository/pg"
)

// Repository provides persistence for promocodes.
type Repository interface {
	Create(ctx context.Context, promo *pg.Promocode) (*pg.Promocode, error)
}
