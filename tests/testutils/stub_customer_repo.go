package testutils

import (
	"context"
	"time"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
)

// stubCustomerRepo is a test implementation of the customer repository.
type StubCustomerRepo struct {
	Ctx context.Context
	// CustomerByTelegramID is returned from FindByTelegramId when set.
	CustomerByTelegramID *domaincustomer.Customer
	Calls                int
}

func (s *StubCustomerRepo) FindById(ctx context.Context, id int64) (*domaincustomer.Customer, error) {
	return nil, nil
}

func (s *StubCustomerRepo) FindByTelegramId(ctx context.Context, telegramId int64) (*domaincustomer.Customer, error) {
	s.Ctx = ctx
	s.Calls++
	if s.CustomerByTelegramID != nil {
		return s.CustomerByTelegramID, nil
	}
	return &domaincustomer.Customer{ID: 1, TelegramID: telegramId, Language: "en", Balance: 0}, nil
}

func (s *StubCustomerRepo) Create(ctx context.Context, c *domaincustomer.Customer) (*domaincustomer.Customer, error) {
	return c, nil
}

func (s *StubCustomerRepo) UpdateFields(ctx context.Context, id int64, updates map[string]interface{}) error {
	return nil
}

func (s *StubCustomerRepo) FindByTelegramIds(ctx context.Context, telegramIDs []int64) ([]domaincustomer.Customer, error) {
	return nil, nil
}

func (s *StubCustomerRepo) DeleteByNotInTelegramIds(ctx context.Context, telegramIDs []int64) error {
	return nil
}

func (s *StubCustomerRepo) CreateBatch(ctx context.Context, customers []domaincustomer.Customer) error {
	return nil
}

func (s *StubCustomerRepo) UpdateBatch(ctx context.Context, customers []domaincustomer.Customer) error {
	return nil
}

func (s *StubCustomerRepo) FindByExpirationRange(ctx context.Context, startDate, endDate time.Time) (*[]domaincustomer.Customer, error) {
	return nil, nil
}
