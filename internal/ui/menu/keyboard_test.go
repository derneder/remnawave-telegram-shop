package menu

import (
	"testing"

	domaincustomer "remnawave-tg-shop-bot/internal/domain/customer"
	"remnawave-tg-shop-bot/internal/pkg/translation"
)

func TestBuildMainKeyboard_AdminButton(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}
	c := &domaincustomer.Customer{}
	kb := BuildMainKeyboard("ru", c, true)
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
	kb = BuildMainKeyboard("ru", c, false)
	for _, r := range kb {
		for _, b := range r {
			if b.CallbackData == CallbackPromoAdminMenu {
				t.Fatalf("admin button should not be shown")
			}
		}
	}
}

func TestBuildPromoRefMain(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}
	userKb := BuildPromoRefMain("ru", false)
	if userKb[0][0].CallbackData != CallbackPromoUserActivate {
		t.Fatalf("unexpected first button %v", userKb[0][0])
	}
	adminKb := BuildPromoRefMain("ru", true)
	if adminKb[0][0].CallbackData != CallbackPromoAdminMenu {
		t.Fatalf("admin panel button missing")
	}
	if len(adminKb) != len(userKb)+1 {
		t.Fatalf("admin menu rows mismatch")
	}
}
