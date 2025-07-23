package customer

import "time"

type Customer struct {
	ID               int64
	TelegramID       int64
	ExpireAt         *time.Time
	CreatedAt        time.Time
	SubscriptionLink *string
	Language         string
	Balance          float64
}
