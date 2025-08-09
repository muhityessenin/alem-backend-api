package http

import (
	"auth/internal/domain"
	"auth/internal/usecase"
	"encoding/json"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

type AuthHandler struct {
	useCase usecase.AuthUseCase
}

func NewAuthHandler(uc usecase.AuthUseCase) *AuthHandler { return &AuthHandler{useCase: uc} }

type registerRequest struct {
	FirstName string      `json:"firstName"`
	LastName  string      `json:"lastName"`
	Email     string      `json:"email"`
	Phone     string      `json:"phone"`
	Password  string      `json:"password"`
	Role      domain.Role `json:"role"`
}
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}
type otpSendRequest struct {
	Phone string `json:"phone"`
}
type otpVerifyRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type tokenResponse struct {
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/auth/register", h.register).Methods("POST")
	router.HandleFunc("/api/auth/login", h.login).Methods("POST")
	router.HandleFunc("/api/auth/refresh", h.refresh).Methods("POST")
	router.HandleFunc("/api/auth/otp/send", h.sendOTP).Methods("POST")
	router.HandleFunc("/api/auth/otp/verify", h.verifyOTP).Methods("POST")

	protected := router.PathPrefix("/api/auth").Subrouter()
	protected.Use(h.JWTMiddleware)
	protected.HandleFunc("/logout", h.logout).Methods("POST")
	protected.HandleFunc("/logout-all", h.logoutAll).Methods("POST")
}

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeErr(w, http.StatusBadRequest, "VALIDATION_ERROR", "email and password are required")
		return
	}
	if req.Role == "" {
		req.Role = domain.StudentRole
	}
	user, pair, err := h.useCase.Register(r.Context(), req.FirstName, req.LastName, req.Email, req.Phone, req.Password, req.Role)
	if err != nil {
		writeErr(w, http.StatusConflict, "REGISTER_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"data": map[string]any{
			"user":  user,
			"token": tokenResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken},
		},
	})
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}
	ua := r.UserAgent()
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	pair, err := h.useCase.Login(r.Context(), req.Email, req.Password, ua, ip)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid credentials")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    tokenResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken},
	})
}

func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}
	ua := r.UserAgent()
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	pair, err := h.useCase.RefreshTokens(r.Context(), req.RefreshToken, ua, ip)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid refresh token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    tokenResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken},
	})
}

func (h *AuthHandler) sendOTP(w http.ResponseWriter, r *http.Request) {
	var req otpSendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Phone == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "phone required")
		return
	}
	// На MVP это можно no-op, но метод готов
	if err := h.useCase.SendOTP(r.Context(), req.Phone); err != nil {
		writeErr(w, http.StatusInternalServerError, "OTP_SEND_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *AuthHandler) verifyOTP(w http.ResponseWriter, r *http.Request) {
	var req otpVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Phone == "" || req.Code == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "phone and code are required")
		return
	}
	ua := r.UserAgent()
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	pair, err := h.useCase.VerifyOTP(r.Context(), req.Phone, req.Code, ua, ip)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid OTP")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    tokenResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken},
	})
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "refreshToken required")
		return
	}
	if err := h.useCase.Logout(r.Context(), req.RefreshToken); err != nil {
		writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid refresh token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Logged out"})
}

func (h *AuthHandler) logoutAll(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserIDKey).(string)
	_ = h.useCase.LogoutAll(r.Context(), userID)
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func writeErr(w http.ResponseWriter, code int, errCode, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error": map[string]string{
			"code":    errCode,
			"message": msg,
		},
	})
}
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
