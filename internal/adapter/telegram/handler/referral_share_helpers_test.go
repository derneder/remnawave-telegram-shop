package handler

import "testing"

func TestBuildDeepLink(t *testing.T) {
	tests := []struct {
		user string
		id   int64
		want string
	}{
		{"bot", 1, "https://t.me/bot?start=ref_1"},
		{"mybot", 42, "https://t.me/mybot?start=ref_42"},
	}
	for _, tt := range tests {
		if got := buildDeepLink(tt.user, tt.id); got != tt.want {
			t.Fatalf("buildDeepLink(%s,%d)=%s, want %s", tt.user, tt.id, got, tt.want)
		}
	}
}

func TestBuildShareURL(t *testing.T) {
	tests := []struct {
		link string
		text string
		want string
	}{
		{"https://t.me/bot?start=ref_1", "Join", "https://t.me/share/url?text=Join&url=https%3A%2F%2Ft.me%2Fbot%3Fstart%3Dref_1"},
		{"a b", "c d", "https://t.me/share/url?text=c+d&url=a+b"},
	}
	for _, tt := range tests {
		if got := buildShareURL(tt.link, tt.text); got != tt.want {
			t.Fatalf("buildShareURL(%s,%s)=%s, want %s", tt.link, tt.text, got, tt.want)
		}
	}
}
