package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	tokenService := NewTokenService("test123")
	token, err := tokenService.GenerateToken(1, "example@mail.com")

	assert.NoError(t, err)

	testCases := []struct {
		name           string
		token          string
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "success",
			token:          "Bearer " + token,
			wantStatusCode: http.StatusOK,
			wantBody:       "OK",
		},

		{
			name:           "missing auth header",
			token:          "",
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
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "invalid token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			middleware := AuthMiddleware(tokenService)
			handler := middleware(testHandler)

			rrl, _ := http.NewRequest(http.MethodGet, "/", nil)
			rrl.Header.Add("Authorization", tc.token)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, rrl)

			assert.Equal(t, tc.wantStatusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.wantBody)

		})
	}
}
