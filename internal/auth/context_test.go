package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserFromContext(t *testing.T) {

	testCases := []struct{
		name string
		claims *Claims
		wantErr error
	}{
		{
			name: "success",
			claims: &Claims{
				UserID: 1,
				Email: "example@mail.com",
			},
			wantErr: nil,
		},

		{
			name: "claims error",
			claims: nil,
			wantErr: ValueNotFound,
		},

	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			ctxWithValue := context.WithValue(ctx, UserContextKey, tc.claims)

			claims, err := GetUserFromContext(ctxWithValue)
			assert.ErrorIs(t, tc.wantErr, err)
			assert.Equal(t, tc.claims, claims)

		})
	}
}