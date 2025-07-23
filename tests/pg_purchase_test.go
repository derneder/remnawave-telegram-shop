package tests

import (
	"testing"

	"remnawave-tg-shop-bot/internal/repository/pg"
)

func TestNewPurchaseRepository(t *testing.T) {
	_ = pg.NewPurchaseRepository(nil)
}
