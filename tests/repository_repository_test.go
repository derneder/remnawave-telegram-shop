package tests

import (
	"context"
	"testing"
	"time"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/repository"
)

// stubCustomerRepo2 implements CustomerRepository for compile-time check
type stubCustomerRepo2 struct{}

func (stubCustomerRepo2) FindById(context.Context, int64) (*domaincustomer.Customer, error) {
	return nil, nil
}
func (stubCustomerRepo2) FindByTelegramId(context.Context, int64) (*domaincustomer.Customer, error) {
	return nil, nil
}
func (stubCustomerRepo2) Create(context.Context, *domaincustomer.Customer) (*domaincustomer.Customer, error) {
	return nil, nil
}
func (stubCustomerRepo2) UpdateFields(context.Context, int64, map[string]interface{}) error {
	return nil
}
func (stubCustomerRepo2) FindByTelegramIds(context.Context, []int64) ([]domaincustomer.Customer, error) {
	return nil, nil
}
func (stubCustomerRepo2) DeleteByNotInTelegramIds(context.Context, []int64) error      { return nil }
func (stubCustomerRepo2) CreateBatch(context.Context, []domaincustomer.Customer) error { return nil }
func (stubCustomerRepo2) UpdateBatch(context.Context, []domaincustomer.Customer) error { return nil }
func (stubCustomerRepo2) FindByExpirationRange(context.Context, time.Time, time.Time) (*[]domaincustomer.Customer, error) {
	return nil, nil
}

func TestInterfaceCompliance(t *testing.T) {
	var _ repository.CustomerRepository = stubCustomerRepo2{}
}
