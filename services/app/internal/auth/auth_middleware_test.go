package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	tokenService := NewTokenService("test123")
	reqBody := "test message"
	token, err := tokenService.GenerateToken(1, "example@mail.com")

	assert.NoError(t, err)

	testCases := []struct {
		name           string
		token          string
		permissive     bool
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "success",
			token:          "Bearer " + token,
			permissive:     false,
			wantStatusCode: http.StatusOK,
			wantBody:       reqBody,
		},

		{
			name:           "missing auth header",
			token:          "",
			permissive:     false,
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "authorization header required",
		},

		{
			name:           "missing 'bearer' in auth header",
			token:          token,
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "invalid authorization format",
		},

		{
			name:           "tampering the token",
			token:          "Bearer invalid",
			permissive:     false,
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "invalid token",
		},

		{
			name:           "permissive",
			token:          "Bearer " + token,
			permissive:     true,
			wantStatusCode: http.StatusOK,
			wantBody:       reqBody,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(reqBody))
			})

			middleware := AuthMiddleware(tokenService, tc.permissive)
			handler := middleware(testHandler)

			rrl, _ := http.NewRequest(http.MethodGet, "/", nil)

			if !tc.permissive {
				rrl.Header.Add("Authorization", tc.token)
			}

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, rrl)

			assert.Equal(t, tc.wantStatusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.wantBody)

		})
	}
}
