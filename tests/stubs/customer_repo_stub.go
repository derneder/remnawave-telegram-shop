package stubs

import (
	"context"
	"time"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
)

// stubCustomerRepo provides empty implementations of all methods
// of the customer repository interface. It is used only for
// compile-time interface compliance checks in tests.
type stubCustomerRepo struct{}

// StubCustomerRepo is an exported alias of stubCustomerRepo so tests in
// other packages can use it without exposing method implementations.
type StubCustomerRepo = stubCustomerRepo

func (stubCustomerRepo) FindById(context.Context, int64) (*domaincustomer.Customer, error) {
	return nil, nil
}

func (stubCustomerRepo) FindByTelegramId(context.Context, int64) (*domaincustomer.Customer, error) {
	return nil, nil
}

func (stubCustomerRepo) Create(context.Context, *domaincustomer.Customer) (*domaincustomer.Customer, error) {
	return nil, nil
}

func (stubCustomerRepo) UpdateFields(context.Context, int64, map[string]interface{}) error {
	return nil
}

func (stubCustomerRepo) FindByTelegramIds(context.Context, []int64) ([]domaincustomer.Customer, error) {
	return nil, nil
}

func (stubCustomerRepo) DeleteByNotInTelegramIds(context.Context, []int64) error { return nil }

func (stubCustomerRepo) CreateBatch(context.Context, []domaincustomer.Customer) error { return nil }

func (stubCustomerRepo) UpdateBatch(context.Context, []domaincustomer.Customer) error { return nil }

func (stubCustomerRepo) FindByExpirationRange(context.Context, time.Time, time.Time) (*[]domaincustomer.Customer, error) {
	return nil, nil
}
