package main

import (
	"testing"

	"remnawave-tg-shop-bot/tests/testutils"
)

func TestMainSmoke(t *testing.T) {
	testutils.SetTestEnv(t)
	t.Setenv("PRICE_1", "10")
	t.Setenv("PRICE_3", "30")
	t.Setenv("PRICE_6", "50")
	t.Setenv("REMNAWAVE_URL", "http://x")
	t.Setenv("DATABASE_URL", "bad://")
	t.Setenv("TRAFFIC_LIMIT", "100")
	main()
}
