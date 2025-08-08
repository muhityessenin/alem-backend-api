package http

import (
	"auth/internal/domain"
	"auth/internal/usecase"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type AuthHandler struct {
	useCase usecase.AuthUseCase
}

func NewAuthHandler(uc usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{useCase: uc}
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) RegisterRoutes(router *mux.Router) {
	// Открытые роуты
	router.HandleFunc("/api/auth/register", h.register).Methods("POST")
	router.HandleFunc("/api/auth/login", h.login).Methods("POST")
	router.HandleFunc("/api/auth/refresh", h.refresh).Methods("POST")

	// Группа защищенных роутов
	protected := router.PathPrefix("/api/auth").Subrouter()
	protected.Use(h.JWTMiddleware) // Применяем middleware ко всей группе
	protected.HandleFunc("/logout", h.logout).Methods("POST")
}

type registerRequest struct {
	Email    string      `json:"email"`
	Password string      `json:"password"`
	Name     string      `json:"name"`
	Role     domain.Role `json:"role"`
}
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Success bool `json:"success"`
	Data    struct {
		AccessToken  string `json:"token"`
		RefreshToken string `json:"refreshToken"`
	} `json:"data"`
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.useCase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		// Здесь можно добавить более детальную обработку ошибок из usecase
		// например, usecase.ErrUserNotFound -> http.StatusNotFound
		http.Error(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Invalid credentials"}}`, http.StatusUnauthorized)
		return
	}

	response := loginResponse{
		Success: true,
	}
	response.Data.AccessToken = accessToken
	response.Data.RefreshToken = refreshToken

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Log the request decoding error
		log.Printf("WARN: failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Здесь нужна валидация, добавим позже

	user, token, err := h.useCase.Register(r.Context(), req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		// The use case has already logged the specific error,
		// this log confirms the handler received an error.
		log.Printf("ERROR: registration failed for email %s: %v", req.Email, err)
		// Here you could check for specific error types, e.g., duplicate email
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user":  user,
			"token": token,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.useCase.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Invalid refresh token"}}`, http.StatusUnauthorized)
		return
	}

	response := loginResponse{Success: true}
	response.Data.AccessToken = accessToken
	response.Data.RefreshToken = refreshToken

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	// В stateless-архитектуре с JWT, реальный выход происходит на клиенте
	// путем удаления токенов. Сервер просто подтверждает операцию.
	// userID, _ := r.Context().Value(UserIDKey).(string) // Можем получить ID, если нужно

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true, "message": "Logged out successfully"}`))
}
