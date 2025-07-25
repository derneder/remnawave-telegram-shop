package handler

import (
	"context"
	"testing"

	"remnawave-tg-shop-bot/internal/pkg/translation"
	referralrepo "remnawave-tg-shop-bot/internal/repository/referral"
	testutils "remnawave-tg-shop-bot/tests"
)

type stubReferralRepo struct{}

func (stubReferralRepo) Create(ctx context.Context, referrerID, refereeID int64) error { return nil }
func (stubReferralRepo) FindByReferee(ctx context.Context, refereeID int64) (*referralrepo.Model, error) {
	return nil, nil
}

func TestNewHandlerReferral(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(nil, nil, tm, &testutils.StubCustomerRepo{}, nil, &stubReferralRepo{}, nil, nil, nil, nil)
	if h == nil {
		t.Fatal("handler is nil")
	}
}
