package cookies

import (
	"context"
	"github.com/romanp1989/gophermart/internal/domain"
)

type ctxAuthKey string

const AuthKey ctxAuthKey = "Token"

func Context(parent context.Context, userID domain.UserID) context.Context {
	return context.WithValue(parent, AuthKey, userID)
}

func UIDFromContext(ctx context.Context) (*domain.UserID, bool) {
	val, ok := ctx.Value(AuthKey).(domain.UserID)
	if !ok || val == 0 {
		return nil, false
	}

	uid := val
	return &uid, true
}
