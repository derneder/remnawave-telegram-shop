package tests

import (
	"context"
	"testing"
	"time"

	"remnawave-tg-shop-bot/internal/pkg/cache"
)

func TestCacheTTL(t *testing.T) {
	c := cache.NewCache(context.Background(), 10*time.Millisecond)
	defer c.Close()

	c.Set(1, 42)
	if v, ok := c.Get(1); !ok || v != 42 {
		t.Fatalf("get expected 42, got %d ok=%v", v, ok)
	}
	time.Sleep(15 * time.Millisecond)
	if _, ok := c.Get(1); ok {
		t.Fatal("value should be expired")
	}
}

func TestCacheDelete(t *testing.T) {
	c := cache.NewCache(context.Background(), time.Minute)
	defer c.Close()

	c.Set(2, 7)
	c.Delete(2)
	if _, ok := c.Get(2); ok {
		t.Fatal("value should be deleted")
	}
}

func TestCacheClose(t *testing.T) {
	c := cache.NewCache(context.Background(), time.Millisecond)
	c.Close()
	// second call should not panic
	c.Close()
}
