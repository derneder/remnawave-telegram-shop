package contextkey

import (
	"context"
	"strings"
)

type contextKey string

const Username contextKey = "username"
const IsAdminKey contextKey = "is_admin"

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

func IsAdminFromContext(ctx context.Context) bool {
	v, ok := ctx.Value(IsAdminKey).(bool)
	return ok && v
}
