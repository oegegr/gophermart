package auth

import (
	"context"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
)

// Authenticator middleware проверяет JWT токен
func Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, claims, err := jwtauth.FromContext(r.Context())
		
		if err != nil || token == nil || !token.Valid{
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Добавляем информацию о пользователе в контекст
		userID, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext извлекает ID пользователя из контекста
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("user_id").(string)
	return userID, ok
}