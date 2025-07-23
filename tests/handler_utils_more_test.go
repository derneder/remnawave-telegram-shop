package tests

import (
	"strconv"
	"strings"
	"testing"
)

func parseCallbackData(data string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(data, "?")
	if len(parts) < 2 {
		return result
	}
	params := strings.Split(parts[1], "&")
	for _, param := range params {
		kv := strings.SplitN(param, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

func buildPaymentBackData(month int, amount int) string {
	if month == 0 {
		return "topup_method?amount=" + strconv.Itoa(amount)
	}
	return "sell?month=" + strconv.Itoa(month) + "&amount=" + strconv.Itoa(amount)
}

func TestParseCallbackData(t *testing.T) {
	res := parseCallbackData("cmd?foo=bar&baz=1")
	if res["foo"] != "bar" || res["baz"] != "1" {
		t.Fatalf("unexpected map %#v", res)
	}
	if len(parseCallbackData("cmd")) != 0 {
		t.Fatalf("expected empty map")
	}
}

func TestBuildPaymentBackData(t *testing.T) {
	if got := buildPaymentBackData(0, 10); got != "topup_method?amount=10" {
		t.Fatalf("wrong back data %s", got)
	}
	if got := buildPaymentBackData(3, 20); got != "sell?month=3&amount=20" {
		t.Fatalf("wrong back data %s", got)
	}
}
