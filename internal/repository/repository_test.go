package repository

import (
	"context"
	"testing"
	"time"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
)

// stubRepo implements CustomerRepository for compile-time check
type stubRepo struct{}

func (stubRepo) FindById(context.Context, int64) (*domaincustomer.Customer, error) { return nil, nil }
func (stubRepo) FindByTelegramId(context.Context, int64) (*domaincustomer.Customer, error) {
	return nil, nil
}
func (stubRepo) Create(context.Context, *domaincustomer.Customer) (*domaincustomer.Customer, error) {
	return nil, nil
}
func (stubRepo) UpdateFields(context.Context, int64, map[string]interface{}) error { return nil }
func (stubRepo) FindByTelegramIds(context.Context, []int64) ([]domaincustomer.Customer, error) {
	return nil, nil
}
func (stubRepo) DeleteByNotInTelegramIds(context.Context, []int64) error      { return nil }
func (stubRepo) CreateBatch(context.Context, []domaincustomer.Customer) error { return nil }
func (stubRepo) UpdateBatch(context.Context, []domaincustomer.Customer) error { return nil }
func (stubRepo) FindByExpirationRange(context.Context, time.Time, time.Time) (*[]domaincustomer.Customer, error) {
	return nil, nil
}

func TestInterfaceCompliance(t *testing.T) {
	var _ CustomerRepository = stubRepo{}
}
