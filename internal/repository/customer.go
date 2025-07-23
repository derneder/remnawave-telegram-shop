package repository

import (
	"context"
	"remnawave-tg-shop-bot/internal/domain/customer"
	"time"
)

type CustomerRepository interface {
	FindById(ctx context.Context, id int64) (*customer.Customer, error)
	FindByTelegramId(ctx context.Context, telegramId int64) (*customer.Customer, error)
	Create(ctx context.Context, c *customer.Customer) (*customer.Customer, error)
	UpdateFields(ctx context.Context, id int64, updates map[string]interface{}) error
	FindByTelegramIds(ctx context.Context, telegramIDs []int64) ([]customer.Customer, error)
	DeleteByNotInTelegramIds(ctx context.Context, telegramIDs []int64) error
	CreateBatch(ctx context.Context, customers []customer.Customer) error
	UpdateBatch(ctx context.Context, customers []customer.Customer) error
	FindByExpirationRange(ctx context.Context, startDate, endDate time.Time) (*[]customer.Customer, error)
}
