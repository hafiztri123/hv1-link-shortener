package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	testCases := []struct {
		name        string
		err         error
		wantErrBody string
	}{
		{
			name:        "Email already exists",
			err:         EmailAlreadyExists,
			wantErrBody: "Email already exists",
		},
		{
			name:        "User not found",
			err:         UserNotFound,
			wantErrBody: "User not found",
		},
		{
			name:        "Invalid credentials",
			err:         InvalidCredentials,
			wantErrBody: "Invalid credentials",
		},
		{
			name:        "Unexpected error",
			err:         UnexpectedError,
			wantErrBody: "Unexpected error has occured",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.err
			assert.Equal(t, tc.wantErrBody, err.Error())
		})
	}
}
