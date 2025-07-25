package referral_test

import (
	"testing"

	referralpg "remnawave-tg-shop-bot/internal/repository/pg/referral"
)

func TestNewRepository(t *testing.T) {
	_ = referralpg.New(nil)
}
