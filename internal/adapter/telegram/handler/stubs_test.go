package handler

import (
	"context"
	"time"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
)

// stubCustomerRepo is a test implementation of the customer repository.
type stubCustomerRepo struct {
	ctx context.Context
	// customerByTelegramID is returned from FindByTelegramId when set.
	customerByTelegramID *domaincustomer.Customer
}

func (s *stubCustomerRepo) FindById(ctx context.Context, id int64) (*domaincustomer.Customer, error) {
	return nil, nil
}

func (s *stubCustomerRepo) FindByTelegramId(ctx context.Context, telegramId int64) (*domaincustomer.Customer, error) {
	s.ctx = ctx
	if s.customerByTelegramID != nil {
		return s.customerByTelegramID, nil
	}
	return &domaincustomer.Customer{ID: 1, TelegramID: telegramId, Language: "en", Balance: 0}, nil
}

func (s *stubCustomerRepo) Create(ctx context.Context, c *domaincustomer.Customer) (*domaincustomer.Customer, error) {
	return c, nil
}

func (s *stubCustomerRepo) UpdateFields(ctx context.Context, id int64, updates map[string]interface{}) error {
	return nil
}

func (s *stubCustomerRepo) FindByTelegramIds(ctx context.Context, telegramIDs []int64) ([]domaincustomer.Customer, error) {
	return nil, nil
}

func (s *stubCustomerRepo) DeleteByNotInTelegramIds(ctx context.Context, telegramIDs []int64) error {
	return nil
}

func (s *stubCustomerRepo) CreateBatch(ctx context.Context, customers []domaincustomer.Customer) error {
	return nil
}

func (s *stubCustomerRepo) UpdateBatch(ctx context.Context, customers []domaincustomer.Customer) error {
	return nil
}

func (s *stubCustomerRepo) FindByExpirationRange(ctx context.Context, startDate, endDate time.Time) (*[]domaincustomer.Customer, error) {
	return nil, nil
}
