package observability_test

import (
	"context"
	"errors"
	"testing"

	"remnawave-tg-shop-bot/internal/observability"
)

func TestMeasure(t *testing.T) {
	called := false
	h := observability.Measure("test", func(ctx context.Context) error {
		called = true
		return nil
	})
	if err := h(context.Background()); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !called {
		t.Fatalf("handler not called")
	}
}

func TestMeasureError(t *testing.T) {
	wantErr := errors.New("fail")
	h := observability.Measure("test", func(ctx context.Context) error { return wantErr })
	if err := h(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("unexpected err %v", err)
	}
}
