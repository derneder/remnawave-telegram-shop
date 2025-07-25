package tests

import (
	"testing"

	"remnawave-tg-shop-bot/internal/repository"
	"remnawave-tg-shop-bot/tests/stubs"
)

func TestInterfaceCompliance(t *testing.T) {
	var _ repository.CustomerRepository = stubs.StubCustomerRepo{}
}
