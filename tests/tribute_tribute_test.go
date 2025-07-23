package tests

import "testing"

func convertPeriodToMonths(period string) int {
	switch period {
	case "monthly":
		return 1
	case "quarterly", "3-month", "3months", "3-months", "q":
		return 3
	case "halfyearly":
		return 6
	default:
		return 1
	}
}

func TestConvertPeriodToMonths(t *testing.T) {
	cases := map[string]int{
		"monthly":    1,
		"quarterly":  3,
		"3-month":    3,
		"3months":    3,
		"3-months":   3,
		"q":          3,
		"halfyearly": 6,
		"unknown":    1,
	}
	for in, want := range cases {
		if got := convertPeriodToMonths(in); got != want {
			t.Errorf("%s => %d want %d", in, got, want)
		}
	}
}
