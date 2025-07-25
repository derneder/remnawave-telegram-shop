package tests

import (
	"fmt"
	"testing"

	"remnawave-tg-shop-bot/internal/pkg/translation"
)

func TestPromoBalanceAppliedText(t *testing.T) {
	tm := translation.GetInstance()
	if err := tm.InitDefaultTranslations(); err != nil {
		t.Fatalf("init translations: %v", err)
	}
	got := fmt.Sprintf(tm.GetText("ru", "promo_balance_applied"), 100, 200)
	want := "✅ Промокод активирован!\nНа ваш баланс зачислено: 100 ₽\nТекущий баланс: 200 ₽"
	if got != want {
		t.Fatalf("unexpected text: %q", got)
	}
}
