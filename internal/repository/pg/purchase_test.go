package pg

import "testing"

func TestNewPurchaseRepository(t *testing.T) {
	_ = NewPurchaseRepository(nil)
}
