package customer

import (
	"context"
	"time"
)

// Repository defines access methods for Customer entities.
type Repository interface {
	FindById(ctx context.Context, id int64) (*Customer, error)
	FindByTelegramId(ctx context.Context, telegramId int64) (*Customer, error)
	Create(ctx context.Context, c *Customer) (*Customer, error)
	UpdateFields(ctx context.Context, id int64, updates map[string]interface{}) error
	FindByTelegramIds(ctx context.Context, telegramIDs []int64) ([]Customer, error)
	DeleteByNotInTelegramIds(ctx context.Context, telegramIDs []int64) error
	CreateBatch(ctx context.Context, customers []Customer) error
	UpdateBatch(ctx context.Context, customers []Customer) error
	FindByExpirationRange(ctx context.Context, startDate, endDate time.Time) (*[]Customer, error)
}
