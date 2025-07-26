package promotion

import (
	"context"
	"crypto/rand"

	"remnawave-tg-shop-bot/internal/repository/pg"
)

const codeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Service handles admin promocodes.
type Creator interface {
	CreateSubscription(ctx context.Context, code string, days, limit int, createdBy int64) (string, error)
	CreateBalance(ctx context.Context, amount, limit int, createdBy int64) (string, error)
	Freeze(ctx context.Context, id int64) error
	Unfreeze(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateSubscription stores subscription promo code with given code and days.
func (s *Service) CreateSubscription(ctx context.Context, code string, days, limit int, createdBy int64) (string, error) {
	if code == "" {
		var err error
		code, err = generateCode()
		if err != nil {
			return "", err
		}
	}
	_, err := s.repo.Create(ctx, &pg.Promocode{
		Code:      code,
		Months:    0,
		Type:      1,
		Days:      days,
		Amount:    0,
		UsesLeft:  limit,
		CreatedBy: createdBy,
		Active:    true,
	})
	if err != nil {
		return "", err
	}
	return code, nil
}

// CreateBalance generates random code and stores it with amount in cents.
func (s *Service) CreateBalance(ctx context.Context, amount, limit int, createdBy int64) (string, error) {
	code, err := generateCode()
	if err != nil {
		return "", err
	}
	_, err = s.repo.Create(ctx, &pg.Promocode{
		Code:      code,
		Type:      2,
		Amount:    amount,
		UsesLeft:  limit,
		CreatedBy: createdBy,
		Active:    true,
	})
	if err != nil {
		return "", err
	}
	return code, nil
}

// Freeze sets promocode status to inactive.
func (s *Service) Freeze(ctx context.Context, id int64) error {
	return s.repo.UpdateStatus(ctx, id, false)
}

// Unfreeze sets promocode status to active.
func (s *Service) Unfreeze(ctx context.Context, id int64) error {
	return s.repo.UpdateStatus(ctx, id, true)
}

// Delete marks promocode as deleted.
func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.UpdateDeleteStatus(ctx, id, true)
}

func generateCode() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = codeAlphabet[int(b[i])%len(codeAlphabet)]
	}
	return string(b), nil
}
