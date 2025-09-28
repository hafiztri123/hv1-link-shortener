package auth

import (
	"context"
	"hafiztri123/app-link-shortener/internal/response"
	"log/slog"
	"net/http"
	"strings"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(ts *TokenService, permissive bool) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if permissive && authHeader == "" {
				h.ServeHTTP(w, r)
				return
			}

			if authHeader == "" {
				slog.Error("missing authorization header")
				response.Error(w, http.StatusUnauthorized, "authorization header required")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			//Authorization must have the prefix "Bearer"
			if tokenString == authHeader {
				slog.Error("missing 'Bearer' in auth header")
				response.Error(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			claims, err := ts.ValidateToken(tokenString)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
