package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func MaskHalfInt(input int) string {
	return MaskHalf(strconv.Itoa(input))
}

func MaskHalfInt64(input int64) string {
	return MaskHalf(strconv.FormatInt(input, 10))
}

func MaskHalf(input string) string {
	if input == "" {
		return input
	}
	if len(input) < 2 {
		return input
	}
	length := len(input)
	visibleLength := length / 2
	maskedLength := length - visibleLength
	return input[:visibleLength] + strings.Repeat("*", maskedLength)
}

func FormatGB(bytes float64) string {
	return fmt.Sprintf("%.2f GB", bytes/(1024*1024*1024))
}

func FormatPrice(amount int) string {
	s := strconv.Itoa(amount)
	n := len(s)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (n-i)%3 == 0 {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}
