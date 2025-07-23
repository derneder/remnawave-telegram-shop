package contextkey

import (
	"context"
	"strings"
)

type contextKey string

const Username contextKey = "username"

func CleanUsername(username string) string {
	return strings.TrimPrefix(strings.TrimSpace(username), "@")
}

func UsernameFromContext(ctx context.Context) string {
	v, ok := ctx.Value(Username).(string)
	if !ok {
		return ""
	}
	return v
}
