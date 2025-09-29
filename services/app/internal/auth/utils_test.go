package auth

import (
	"context"
	"hafiztri123/app-link-shortener/internal/shared"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserFromContext(t *testing.T) {

	testCases := []struct {
		name    string
		claims  *Claims
		wantErr error
	}{
		{
			name: "success",
			claims: &Claims{
				UserID: 1,
				Email:  "example@mail.com",
			},
			wantErr: nil,
		},

		{
			name:    "claims error",
			claims:  nil,
			wantErr: ValueNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			ctxWithValue := context.WithValue(ctx, shared.UserContextKey, tc.claims)

			claims, err := GetUserFromContext(ctxWithValue)
			assert.ErrorIs(t, tc.wantErr, err)
			assert.Equal(t, tc.claims, claims)

		})
	}
}

func TestValueNotFoundErr_ErrorMethod(t *testing.T) {
	err := &ValueNotFoundErr{Action: "test action"}
	result := err.Error()
	expected := "Unexpected error has occured, please try again"
	assert.Equal(t, expected, result)
}

func TestValueNotFoundErr_IsMethod(t *testing.T) {
	err1 := &ValueNotFoundErr{Action: "action1"}
	err2 := &ValueNotFoundErr{Action: "action2"}

	assert.True(t, err1.Is(err2))

	otherErr := assert.AnError
	assert.False(t, err1.Is(otherErr))
}
