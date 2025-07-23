package customer_test

import (
	"context"
	"testing"
	"time"

	domain "remnawave-tg-shop-bot/internal/domain/customer"
	svc "remnawave-tg-shop-bot/internal/service/customer"
)

type stubRepo struct{}

func (stubRepo) FindById(context.Context, int64) (*domain.Customer, error)          { return nil, nil }
func (stubRepo) FindByTelegramId(context.Context, int64) (*domain.Customer, error)  { return nil, nil }
func (stubRepo) Create(context.Context, *domain.Customer) (*domain.Customer, error) { return nil, nil }
func (stubRepo) UpdateFields(context.Context, int64, map[string]interface{}) error  { return nil }
func (stubRepo) FindByTelegramIds(context.Context, []int64) ([]domain.Customer, error) {
	return nil, nil
}
func (stubRepo) DeleteByNotInTelegramIds(context.Context, []int64) error { return nil }
func (stubRepo) CreateBatch(context.Context, []domain.Customer) error    { return nil }
func (stubRepo) UpdateBatch(context.Context, []domain.Customer) error    { return nil }
func (stubRepo) FindByExpirationRange(context.Context, time.Time, time.Time) (*[]domain.Customer, error) {
	return nil, nil
}

func TestAlias(t *testing.T) {
	var _ svc.Repository = stubRepo{}
}
