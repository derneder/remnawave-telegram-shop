package menu

import (
	"testing"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

func TestBuildLKMenu_AdminButton(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}
	c := &domaincustomer.Customer{}
	kb := BuildLKMenu("ru", c, true)
	found := false
	for _, r := range kb {
		for _, b := range r {
			if b.CallbackData == CallbackPromoAdminMenu {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("admin button missing")
	}
	kb = BuildLKMenu("ru", c, false)
	for _, r := range kb {
		for _, b := range r {
			if b.CallbackData == CallbackPromoAdminMenu {
				t.Fatalf("admin button should not be shown")
			}
		}
	}
}

func TestBuildPromoRefMenu(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}
	kb := BuildPromoRefMenu("ru")
	for _, row := range kb {
		for _, b := range row {
			if b.CallbackData == CallbackPromoAdminMenu || b.Text == tm.GetText("ru", "faq_button") {
				t.Fatalf("unexpected button %v", b)
			}
		}
	}
}
