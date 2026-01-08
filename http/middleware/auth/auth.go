package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type User struct {
	Name  string
	Email string
}

type contextKey string

const userContextKey contextKey = "user"

func ContextUser(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey).(*User)
	return user, ok
}

var ErrInvalidToken = errors.New("invalid token")

type TokenChecker interface {
	CheckToken(ctx context.Context, token string) (*User, error)
}

func CheckAuth(checker TokenChecker) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]
			user, err := checker.CheckToken(r.Context(), token)
			if err != nil {
				if errors.Is(err, ErrInvalidToken) {
					http.Error(w, "Invalid token", http.StatusUnauthorized)
				} else {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
