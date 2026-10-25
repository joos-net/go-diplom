package middleware

import (
	"advanced-blog-management-system/pkg/auth"
	"context"
	"errors"
	"net/http"
	"strings"
)

// UserContextKey используется для сохранения данных пользователя в context
type UserContextKey string

const UserKey UserContextKey = "user"

type JWTSecretProvider func() string

// AuthMiddleware проверяет JWT токен и добавляет данные пользователя в context
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := ExtractToken(r)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			claims, err := auth.ValidateToken(token, secret)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), UserKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware проверяет JWT токен если он присутствует, но не обязателен
func OptionalAuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := ExtractToken(r)
			if err == nil {
				if claims, err := auth.ValidateToken(token, secret); err == nil {
					ctx := context.WithValue(r.Context(), UserKey, claims.UserID)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext извлекает пользователя из context
func GetUserIDFromContext(r *http.Request) (int, bool) {
	v := r.Context().Value(UserKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int)
	return id, ok
}

// ExtractToken извлекает JWT токен из заголовка Authorization
func ExtractToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", errors.New("missing authorization header")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization format")
	}
	return parts[1], nil
}
