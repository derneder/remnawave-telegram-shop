package referral

import (
	"context"
	"time"
)

// Repository defines operations for referral storage.
type Repository interface {
	Create(ctx context.Context, referrerID int64, refereeID int64) error
	FindByReferee(ctx context.Context, refereeID int64) (*Model, error)
	MarkBonusGranted(ctx context.Context, referralID int64) error
}

// Model represents a referral record.
type Model struct {
	ID           int64
	ReferrerID   int64
	RefereeID    int64
	CreatedAt    time.Time
	BonusGranted bool
}
