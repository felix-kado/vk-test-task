package middleware

import (
	"context"
	"net/http"
	"strings"

	"example.com/market/internal/domain"
)

type contextKey string

// UserIDKey is the key for the user ID in the context.
const UserIDKey contextKey = "userID"

// TokenParser defines the interface for parsing a token.
type TokenParser interface {
	ParseToken(ctx context.Context, token string) (*domain.User, error)
}

// AuthCtx is a middleware that extracts the JWT from the Authorization header
// and sets the user information in the request context.
func AuthCtx(parser TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				http.Error(w, "invalid auth header", http.StatusUnauthorized)
				return
			}

			tokenStr := headerParts[1]
			user, err := parser.ParseToken(r.Context(), tokenStr)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, user.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
