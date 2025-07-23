package tests

import (
	"context"
	"testing"
	"time"

	domain "remnawave-tg-shop-bot/internal/domain/customer"
	svc "remnawave-tg-shop-bot/internal/service/customer"
)

type stubCustomerRepoAlias struct{}

func (stubCustomerRepoAlias) FindById(context.Context, int64) (*domain.Customer, error) {
	return nil, nil
}
func (stubCustomerRepoAlias) FindByTelegramId(context.Context, int64) (*domain.Customer, error) {
	return nil, nil
}
func (stubCustomerRepoAlias) Create(context.Context, *domain.Customer) (*domain.Customer, error) {
	return nil, nil
}
func (stubCustomerRepoAlias) UpdateFields(context.Context, int64, map[string]interface{}) error {
	return nil
}
func (stubCustomerRepoAlias) FindByTelegramIds(context.Context, []int64) ([]domain.Customer, error) {
	return nil, nil
}
func (stubCustomerRepoAlias) DeleteByNotInTelegramIds(context.Context, []int64) error { return nil }
func (stubCustomerRepoAlias) CreateBatch(context.Context, []domain.Customer) error    { return nil }
func (stubCustomerRepoAlias) UpdateBatch(context.Context, []domain.Customer) error    { return nil }
func (stubCustomerRepoAlias) FindByExpirationRange(context.Context, time.Time, time.Time) (*[]domain.Customer, error) {
	return nil, nil
}

func TestAlias(t *testing.T) {
	var _ svc.Repository = stubCustomerRepoAlias{}
}
