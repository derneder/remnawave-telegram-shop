package tests

import (
	"testing"

	"remnawave-tg-shop-bot/translations"
)

func TestEmbeddedFiles(t *testing.T) {
	if _, err := translations.FS.ReadFile("en.yml"); err != nil {
		t.Fatalf("failed to read en.yml: %v", err)
	}
}
