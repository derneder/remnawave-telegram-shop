package tests

import (
	"testing"

	svc "remnawave-tg-shop-bot/internal/service/customer"
)

func TestAlias(t *testing.T) {
	var _ svc.Repository = &StubCustomerRepo{}
}
