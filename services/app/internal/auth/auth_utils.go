package auth

import (
	"context"
	"hafiztri123/app-link-shortener/internal/shared"
)

func GetUserFromContext(ctx context.Context) (*Claims, error) {
	user, ok := ctx.Value(shared.UserContextKey).(*Claims)
	if !ok || user == nil {
		return nil, ValueNotFound
	}
	return user, nil
}
