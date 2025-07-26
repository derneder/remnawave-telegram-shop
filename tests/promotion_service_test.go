package tests

import (
	"context"
	"fmt"
	"testing"

	"remnawave-tg-shop-bot/internal/repository/pg"
	"remnawave-tg-shop-bot/internal/service/promotion"
)

type stubPromoRepo struct {
	store  map[int64]*pg.Promocode
	byCode map[string]int64
	nextID int64
}

func (s *stubPromoRepo) Create(ctx context.Context, p *pg.Promocode) (*pg.Promocode, error) {
	if s.store == nil {
		s.store = make(map[int64]*pg.Promocode)
		s.byCode = make(map[string]int64)
	}
	if _, ok := s.byCode[p.Code]; ok {
		return nil, fmt.Errorf("duplicate")
	}
	s.nextID++
	p.ID = s.nextID
	s.store[p.ID] = p
	s.byCode[p.Code] = p.ID
	return p, nil
}

func (s *stubPromoRepo) UpdateStatus(ctx context.Context, id int64, active bool) error {
	if p, ok := s.store[id]; ok {
		p.Active = active
	}
	return nil
}

func (s *stubPromoRepo) UpdateDeleteStatus(ctx context.Context, id int64, deleted bool) error {
	if p, ok := s.store[id]; ok {
		p.Deleted = deleted
	}
	return nil
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
	if _, err := svc.CreateSubscription(context.Background(), "CODE", 30, 1, 1); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if _, err := svc.CreateSubscription(context.Background(), "CODE", 30, 1, 1); err == nil {
		t.Fatal("expected error for duplicate code")
	}
}

func TestFreezeUnfreezeDelete(t *testing.T) {
	repo := &stubPromoRepo{}
	svc := promotion.NewService(repo)
	code, err := svc.CreateBalance(context.Background(), 10, 1, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	id := repo.byCode[code]
	if err := svc.Freeze(context.Background(), id); err != nil {
		t.Fatalf("freeze: %v", err)
	}
	if repo.store[id].Active {
		t.Fatal("not frozen")
	}
	if err := svc.Unfreeze(context.Background(), id); err != nil {
		t.Fatalf("unfreeze: %v", err)
	}
	if !repo.store[id].Active {
		t.Fatal("not unfrozen")
	}
	if err := svc.Delete(context.Background(), id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !repo.store[id].Deleted {
		t.Fatal("not deleted")
	}
}
