package tests

import (
	"testing"

	"remnawave-tg-shop-bot/internal/repository"
)

func TestInterfaceCompliance(t *testing.T) {
	var _ repository.CustomerRepository = &StubCustomerRepo{}
}
