package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"example.com/market/internal/domain"
)

type contextKey string

// UserIDKey is the key for the user ID in the context.
const UserIDKey contextKey = "userID"

// AuthService defines the interface for authenticating a user.
type AuthService interface {
	ParseToken(ctx context.Context, token string) (*domain.User, error)
}

// AuthCtx is a middleware that extracts the JWT from the Authorization header
// and sets the user information in the request context.
func AuthOptionalCtx(authService AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				next.ServeHTTP(w, r)
				return
			}

			tokenStr := headerParts[1]
			tokenPrefix := tokenStr
			if len(tokenStr) > 20 {
				tokenPrefix = tokenStr[:20] + "..."
			}
			slog.Debug("parsing token in AuthOptionalCtx", slog.String("token_prefix", tokenPrefix))
			
			user, err := authService.ParseToken(r.Context(), tokenStr)
			if err != nil {
				slog.Debug("failed to parse token in AuthOptionalCtx", slog.String("error", err.Error()))
				next.ServeHTTP(w, r) // Proceed without user if token is invalid
				return
			}

			userID := int64(user.ID)
			slog.Debug("successfully parsed token in AuthOptionalCtx", slog.Int64("user_id", userID), slog.String("login", user.Login))
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthCtx(authService AuthService) func(http.Handler) http.Handler {
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
			user, err := authService.ParseToken(r.Context(), tokenStr)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			userID := int64(user.ID)
			slog.Debug("user authenticated, adding user_id to context", slog.Int64("user_id", userID), slog.String("type", fmt.Sprintf("%T", userID)))
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
