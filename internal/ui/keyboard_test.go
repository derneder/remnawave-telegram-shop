package ui

import (
	"remnawave-tg-shop-bot/internal/pkg/translation"
	"testing"
)

func TestMakeConnectButton(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}
	btn := makeConnectButton("", "en")
	if btn.CallbackData != "connect" || btn.URL != "" {
		t.Fatalf("expected callback button, got %#v", btn)
	}

	url := "https://example.com"
	btn = makeConnectButton(url, "en")
	if btn.URL != url || btn.CallbackData != "" {
		t.Fatalf("expected url button, got %#v", btn)
	}
}
