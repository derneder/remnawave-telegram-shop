package tests

import (
	"testing"

	svc "remnawave-tg-shop-bot/internal/service/customer"
	"remnawave-tg-shop-bot/tests/stubs"
)

func TestAlias(t *testing.T) {
	var _ svc.Repository = stubs.StubCustomerRepo{}
}
