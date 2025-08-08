package http

import (
	"context"
	"net/http"
	"strings"

	_ "auth/internal/usecase" // Нам нужны Claims отсюда
)

// CtxKey - тип для ключа в контексте запроса.
type CtxKey string

const (
	// UserIDKey - ключ для хранения ID пользователя в контексте.
	UserIDKey CtxKey = "userID"
	// UserRoleKey - ключ для хранения роли пользователя в контексте.
	UserRoleKey CtxKey = "userRole"
)

// JWTMiddleware проверяет токен авторизации.
func (h *AuthHandler) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Получаем заголовок Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Missing authorization header"}}`, http.StatusUnauthorized)
			return
		}

		// 2. Проверяем формат "Bearer <token>"
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			http.Error(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Invalid authorization header format"}}`, http.StatusUnauthorized)
			return
		}

		tokenString := headerParts[1]

		// 3. Парсим и валидируем токен
		claims, err := h.useCase.ParseToken(r.Context(), tokenString)
		if err != nil {
			http.Error(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Invalid or expired token"}}`, http.StatusUnauthorized)
			return
		}

		// 4. Добавляем данные из токена в контекст запроса
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		// 5. Передаем управление следующему обработчику
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
