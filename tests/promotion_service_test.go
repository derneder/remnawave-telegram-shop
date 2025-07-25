package tests

import (
	"context"
	"fmt"
	"testing"

	"remnawave-tg-shop-bot/internal/repository/pg"
	"remnawave-tg-shop-bot/internal/service/promotion"
)

type stubPromoRepo struct{ store map[string]*pg.Promocode }

func (s *stubPromoRepo) Create(ctx context.Context, p *pg.Promocode) (*pg.Promocode, error) {
	if s.store == nil {
		s.store = make(map[string]*pg.Promocode)
	}
	if _, ok := s.store[p.Code]; ok {
		return nil, fmt.Errorf("duplicate")
	}
	s.store[p.Code] = p
	return p, nil
}

func TestCreateBalanceCodeUnique(t *testing.T) {
	repo := &stubPromoRepo{}
	svc := promotion.NewService(repo)
	c1, err := svc.CreateBalance(context.Background(), 10, 1, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(c1) != 20 {
		t.Fatalf("code length %d", len(c1))
	}
	c2, err := svc.CreateBalance(context.Background(), 10, 1, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if c1 == c2 {
		t.Fatal("codes not unique")
	}
}

func TestCreateSubscriptionDuplicate(t *testing.T) {
	repo := &stubPromoRepo{}
	svc := promotion.NewService(repo)
	if err := svc.CreateSubscription(context.Background(), "CODE", 30, 1, 1); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := svc.CreateSubscription(context.Background(), "CODE", 30, 1, 1); err == nil {
		t.Fatal("expected error for duplicate code")
	}
}
