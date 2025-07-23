package tests

import (
	"testing"

	"remnawave-tg-shop-bot/utils"
)

func TestMaskHalf(t *testing.T) {
	if utils.MaskHalf("abcd") != "ab**" {
		t.Fatalf("unexpected mask")
	}
	if utils.MaskHalf("") != "" {
		t.Fatalf("empty not handled")
	}
}

func TestFormatPrice(t *testing.T) {
	if utils.FormatPrice(1234567) != "1 234 567" {
		t.Fatalf("wrong format")
	}
}
