package tests

import (
	"testing"

	"remnawave-tg-shop-bot/internal/pkg/translation"
	"remnawave-tg-shop-bot/internal/ui"
)

func TestMakeConnectButton(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}
	btn := ui.MakeConnectButton("", "en")
	if btn.CallbackData != "connect" || btn.URL != "" {
		t.Fatalf("expected callback button, got %#v", btn)
	}

	url := "https://example.com"
	btn = ui.MakeConnectButton(url, "en")
	if btn.URL != url || btn.CallbackData != "" {
		t.Fatalf("expected url button, got %#v", btn)
	}
}
