package auth

import "context"

func GetUserFromContext(ctx context.Context) (*Claims, error) {
	user, ok := ctx.Value(UserContextKey).(*Claims)
	if !ok || user == nil {
		return nil, ValueNotFound
	}
	return user, nil
}
