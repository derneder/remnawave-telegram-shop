package tests

import (
	"context"
	"testing"

	"remnawave-tg-shop-bot/internal/pkg/contextkey"
)

func TestCleanUsername(t *testing.T) {
	if got := contextkey.CleanUsername(" @user"); got != "user" {
		t.Fatalf("unexpected clean %q", got)
	}
	if got := contextkey.CleanUsername("@bob"); got != "bob" {
		t.Fatalf("unexpected clean %q", got)
	}
}

func TestUsernameFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextkey.Username, "name")
	if contextkey.UsernameFromContext(ctx) != "name" {
		t.Fatalf("wrong value")
	}
	if contextkey.UsernameFromContext(context.Background()) != "" {
		t.Fatalf("expected empty string")
	}
}
