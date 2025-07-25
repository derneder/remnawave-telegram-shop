package tests

import (
	"context"
	"os"
	"testing"
	"time"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
)

// CtxKey is used in tests for context propagation checks.
type CtxKey struct{}

// SetTestEnv sets common environment variables required for tests.
func SetTestEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DISABLE_ENV_FILE", "true")
	t.Setenv("ADMIN_TELEGRAM_IDS", "1")
	t.Setenv("TELEGRAM_TOKEN", "t")
	t.Setenv("TRIAL_TRAFFIC_LIMIT", "1")
	t.Setenv("TRIAL_DAYS", "1")
	t.Setenv("PRICE_1", "1")
	t.Setenv("PRICE_3", "1")
	t.Setenv("PRICE_6", "1")
	t.Setenv("REMNAWAVE_URL", "http://example.com")
	t.Setenv("REMNAWAVE_TOKEN", "x")
	t.Setenv("DATABASE_URL", "db")
	t.Setenv("TRAFFIC_LIMIT", "1")
	t.Setenv("REFERRAL_DAYS", "0")
	t.Setenv("REFERRAL_BONUS", "0")
	t.Setenv("CRYPTO_PAY_ENABLED", "false")
	t.Setenv("TELEGRAM_STARS_ENABLED", "false")
	t.Setenv("SUBSCRIPTION_ALLOWED_HOSTS", "example.com")
}

// MustGetEnvForTest fails the test if the environment variable is not set.
func MustGetEnvForTest(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Fatalf("%s not set", key)
	}
	return v
}

// StubCustomerRepo is a test implementation of the customer repository.
type StubCustomerRepo struct {
	Ctx                  context.Context
	CustomerByTelegramID *domaincustomer.Customer
	Calls                int
}

func (s *StubCustomerRepo) FindById(ctx context.Context, id int64) (*domaincustomer.Customer, error) {
	return nil, nil
}

func (s *StubCustomerRepo) FindByTelegramId(ctx context.Context, telegramId int64) (*domaincustomer.Customer, error) {
	s.Ctx = ctx
	s.Calls++
	if s.CustomerByTelegramID != nil {
		return s.CustomerByTelegramID, nil
	}
	return &domaincustomer.Customer{ID: 1, TelegramID: telegramId, Language: "en", Balance: 0}, nil
}

func (s *StubCustomerRepo) Create(ctx context.Context, c *domaincustomer.Customer) (*domaincustomer.Customer, error) {
	return c, nil
}

func (s *StubCustomerRepo) UpdateFields(ctx context.Context, id int64, updates map[string]interface{}) error {
	return nil
}

func (s *StubCustomerRepo) FindByTelegramIds(ctx context.Context, telegramIDs []int64) ([]domaincustomer.Customer, error) {
	return nil, nil
}

func (s *StubCustomerRepo) DeleteByNotInTelegramIds(ctx context.Context, telegramIDs []int64) error {
	return nil
}

func (s *StubCustomerRepo) CreateBatch(ctx context.Context, customers []domaincustomer.Customer) error {
	return nil
}

func (s *StubCustomerRepo) UpdateBatch(ctx context.Context, customers []domaincustomer.Customer) error {
	return nil
}

func (s *StubCustomerRepo) FindByExpirationRange(ctx context.Context, startDate, endDate time.Time) (*[]domaincustomer.Customer, error) {
	return nil, nil
}
