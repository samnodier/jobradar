package auth

import (
	"context"
)

type contextKey string

const sessionKey contextKey = "session"

func SessionFromContext(ctx context.Context) (*Session, bool) {
	session, ok := ctx.Value(sessionKey).(*Session)
	return session, ok
}
