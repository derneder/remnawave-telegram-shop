package handler

import (
	"testing"

	"remnawave-tg-shop-bot/internal/pkg/translation"
	pg "remnawave-tg-shop-bot/internal/repository/pg"
	uimenu "remnawave-tg-shop-bot/internal/ui/menu"
)

func TestBuildPromoItemButtons_Active(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatal(err)
	}
	p := pg.Promocode{ID: 1, Active: true}
	buttons := buildPromoItemButtons("en", p)
	if len(buttons) != 2 {
		t.Fatalf("expected 2 buttons, got %d", len(buttons))
	}
	if buttons[0].CallbackData != uimenu.CallbackPromoMyFreeze+":1" {
		t.Fatalf("freeze button missing")
	}
	if buttons[1].CallbackData != uimenu.CallbackPromoMyDelete+":1" {
		t.Fatalf("delete button missing")
	}
}

func TestBuildPromoItemButtons_Inactive(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatal(err)
	}
	p := pg.Promocode{ID: 2, Active: false}
	buttons := buildPromoItemButtons("en", p)
	if len(buttons) != 2 {
		t.Fatalf("expected 2 buttons, got %d", len(buttons))
	}
	if buttons[0].CallbackData != uimenu.CallbackPromoMyUnfreeze+":2" {
		t.Fatalf("unfreeze button missing")
	}
	if buttons[1].CallbackData != uimenu.CallbackPromoMyDelete+":2" {
		t.Fatalf("delete button missing")
	}
}
