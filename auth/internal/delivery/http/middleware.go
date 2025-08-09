package http

import (
	"context"
	"net/http"
	"strings"

	_ "auth/internal/usecase"
)

type CtxKey string

const (
	UserIDKey   CtxKey = "userID"
	UserRoleKey CtxKey = "userRole"
)

func (h *AuthHandler) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ah := r.Header.Get("Authorization")
		if ah == "" {
			writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header")
			return
		}
		parts := strings.Split(ah, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format")
			return
		}
		claims, err := h.useCase.ParseToken(r.Context(), parts[1])
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
