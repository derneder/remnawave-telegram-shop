package referral

import "context"

// Repository defines persistence operations for referrals.
type Repository interface {
	Create(ctx context.Context, referrerID, refereeID int64) (*Referral, error)
	FindByReferrer(ctx context.Context, referrerID int64) ([]Referral, error)
	CountByReferrer(ctx context.Context, referrerID int64) (int, error)
	FindByReferee(ctx context.Context, refereeID int64) (*Referral, error)
	MarkBonusGranted(ctx context.Context, referralID int64) error
}
