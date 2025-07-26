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

func TestBuildPromoRefMain(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}
	kb := BuildPromoRefMain("ru", false)
	for _, row := range kb {
		for _, b := range row {
			if b.CallbackData == CallbackPromoAdminMenu || b.Text == tm.GetText("ru", "faq_button") {
				t.Fatalf("unexpected button %v", b)
			}
		}
	}
	foundPersonal := false
	for _, row := range kb {
		for _, b := range row {
			if b.CallbackData == CallbackPersonalCodes {
				foundPersonal = true
			}
		}
	}
	if !foundPersonal {
		t.Fatalf("personal codes button missing")
	}
	// ensure admin button is present when requested
	kb = BuildPromoRefMain("ru", true)
	found := false
	for _, row := range kb {
		for _, b := range row {
			if b.CallbackData == CallbackPromoAdminMenu {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("admin button missing")
	}
}

func TestBuildPersonalCodesMenu(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}
	kb := BuildPersonalCodesMenu("en")
	haveCreate := false
	haveList := false
	haveBack := false
	for _, row := range kb {
		for _, b := range row {
			switch b.CallbackData {
			case CallbackPersonalCreate:
				haveCreate = true
			case CallbackPromoMyList:
				haveList = true
			case "referral":
				haveBack = true
			}
		}
	}
	if !(haveCreate && haveList && haveBack) {
		t.Fatalf("buttons missing")
	}
}

func TestBuildAdminPromoMenus(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}

	kb := BuildAdminPromoMenu("en")
	if len(kb) != 2 || kb[0][0].CallbackData != CallbackPromoAdminMenuPromos {
		t.Fatalf("unexpected admin promo menu: %#v", kb)
	}

	kb = BuildAdminPromoCodesMenu("en")
	if len(kb) != 3 ||
		kb[0][0].CallbackData != CallbackPromoAdminSubStart ||
		kb[1][0].CallbackData != CallbackPromoAdminBalanceStart ||
		kb[2][0].CallbackData != CallbackPromoAdminMenu {
		t.Fatalf("unexpected admin promo codes menu: %#v", kb)
	}
}
