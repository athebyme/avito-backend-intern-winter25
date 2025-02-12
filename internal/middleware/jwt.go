package middleware

import (
	"avito-backend-intern-winter25/internal/auth"
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	ContextKeyUserID contextKey = "userID"
)

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Отсутствует токен авторизации", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Неверный формат токена", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Неверный или истёкший токен", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, "Невалидные данные токена", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
