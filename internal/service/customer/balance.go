package customer

import (
	"context"
	"fmt"
)

type Service interface {
	AddBalance(ctx context.Context, telegramID int64, amountRUB int64) error
}

// BalanceService implements Service.
type BalanceService struct {
	repo Repository
}

func NewService(repo Repository) *BalanceService {
	return &BalanceService{repo: repo}
}

func (s *BalanceService) AddBalance(ctx context.Context, telegramID int64, amountRUB int64) error {
	c, err := s.repo.FindByTelegramId(ctx, telegramID)
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("customer %d not found", telegramID)
	}
	newBal := c.Balance + float64(amountRUB)
	return s.repo.UpdateFields(ctx, c.ID, map[string]interface{}{"balance": newBal})
}
