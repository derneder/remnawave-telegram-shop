package handler

import (
	"fmt"
	"net/url"
)

func buildDeepLink(botUsername string, refCode int64) string {
	return fmt.Sprintf("https://t.me/%s?start=ref_%d", botUsername, refCode)
}

func buildShareURL(deepLink, shareText string) string {
	v := url.Values{}
	v.Set("url", deepLink)
	v.Set("text", shareText)
	return "https://t.me/share/url?" + v.Encode()
}
