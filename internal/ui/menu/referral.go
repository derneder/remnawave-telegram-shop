package menu

import "fmt"

// BuildReferralLink returns t.me deep-link for given bot username and referral code.
func BuildReferralLink(botUsername, refCode string) string {
	return fmt.Sprintf("https://t.me/%s?start=%s", botUsername, refCode)
}
