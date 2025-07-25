package config

import "testing"

func TestValidatePath(t *testing.T) {
	if err := ValidatePath("/tribute", "TRIBUTE_WEBHOOK_URL"); err != nil {
		t.Fatalf("valid path: %v", err)
	}
	if err := ValidatePath("tribute", "TRIBUTE_WEBHOOK_URL"); err == nil || err.Error() != "TRIBUTE_WEBHOOK_URL must start with \"/\"" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ValidatePath("", "TRIBUTE_WEBHOOK_URL"); err == nil || err.Error() != "TRIBUTE_WEBHOOK_URL is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}
