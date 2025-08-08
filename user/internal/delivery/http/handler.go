package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log" // <--- Добавлен импорт
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"user/internal/domain"

	"user/internal/usecase"
)

// UserHandler - обработчик для юзер-сервиса.
type UserHandler struct {
	userUseCase  usecase.UserUseCase
	tokenUseCase usecase.TokenUseCase
}

// CtxKey - тип для ключа в контексте запроса.
type CtxKey string

const (
	UserIDKey   CtxKey = "userID"
	UserRoleKey CtxKey = "userRole"
)

func NewUserHandler(uc usecase.UserUseCase, tuc usecase.TokenUseCase) *UserHandler {
	return &UserHandler{userUseCase: uc, tokenUseCase: tuc}
}

func (h *UserHandler) RegisterRoutes(router *mux.Router) {
	protected := router.PathPrefix("/api/users").Subrouter()
	protected.Use(h.JWTMiddleware)
	protected.HandleFunc("/profile", h.getProfile).Methods("GET")
	protected.HandleFunc("/profile", h.updateProfile).Methods("PUT")
	protected.HandleFunc("/upload-avatar", h.uploadAvatar).Methods("POST")

	students := router.PathPrefix("/api/students").Subrouter()
	students.Use(h.JWTMiddleware)
	students.HandleFunc("", h.listStudents).Methods("GET")
}

type updateProfileRequest struct {
	Name          string               `json:"name"`
	Age           int                  `json:"age"`
	Avatar        string               `json:"avatar"`
	LearningGoals []string             `json:"learningGoals"`
	Description   string               `json:"description"`
	Notifications domain.Notifications `json:"notifications"`
}

func (h *UserHandler) getProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		// Эта ошибка не должна происходить, если middleware работает правильно.
		// Логируем ее как серьезную проблему на сервере.
		log.Printf("ERROR: UserID not found in context")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.userUseCase.GetProfile(r.Context(), userID)
	if err != nil {
		// Логируем фактическую ошибку от usecase/repository
		log.Printf("ERROR: Failed to get user profile for userID %s: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("WARN: Missing authorization header from %s", r.RemoteAddr)
			http.Error(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Missing authorization header"}}`, http.StatusUnauthorized)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			log.Printf("WARN: Invalid authorization header format from %s", r.RemoteAddr)
			http.Error(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Invalid authorization header format"}}`, http.StatusUnauthorized)
			return
		}

		tokenString := headerParts[1]

		claims, err := h.tokenUseCase.ParseToken(r.Context(), tokenString)
		if err != nil {
			// Логируем ошибку парсинга токена - это может указывать на попытку взлома или истекший токен.
			log.Printf("WARN: Failed to parse token: %v", err)
			http.Error(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Invalid or expired token"}}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *UserHandler) updateProfile(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из токена, чтобы пользователь мог менять только свой профиль.
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		log.Printf("ERROR: UserID not found in context")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: Failed to decode update profile request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Создаем объект domain.User для передачи в usecase.
	userToUpdate := &domain.User{
		ID:            userID, // Важно: ID берем из токена!
		Name:          req.Name,
		Age:           req.Age,
		Avatar:        req.Avatar,
		LearningGoals: req.LearningGoals,
		Description:   req.Description,
		Notifications: req.Notifications,
	}

	updatedUser, err := h.userUseCase.UpdateProfile(r.Context(), userToUpdate)
	if err != nil {
		log.Printf("ERROR: Failed to update profile for userID %s: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    updatedUser,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из токена
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		log.Printf("ERROR: UserID not found in context")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Устанавливаем максимальный размер файла (например, 10 MB)
	r.ParseMultipartForm(10 << 20)

	// Получаем файл из multipart-формы. "avatar" - это имя поля в форме
	file, handler, err := r.FormFile("avatar")
	if err != nil {
		log.Printf("ERROR: Error retrieving file from form-data: %v", err)
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// --- Логика сохранения файла ---
	// В реальном приложении здесь будет вызов API Cloudflare.
	// Мы же для примера сохраним файл локально.

	// Создаем директорию для загрузок, если ее нет
	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, os.ModePerm)

	// Формируем уникальное имя файла, чтобы избежать коллизий
	fileName := fmt.Sprintf("%s%s", userID, filepath.Ext(handler.Filename))
	filePath := filepath.Join(uploadDir, fileName)

	// Создаем файл на сервере
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("ERROR: Could not create file on server: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Копируем содержимое загруженного файла в созданный файл
	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("ERROR: Could not save file: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// --- Конец логики сохранения файла ---

	// Формируем URL, по которому будет доступен файл
	// (в вашем случае это будет URL от Cloudflare)
	avatarURL := "http://localhost:8082/" + filePath // Пример для локального сервера

	// Обновляем URL аватара в базе данных
	if err := h.userUseCase.UpdateAvatar(r.Context(), userID, avatarURL); err != nil {
		log.Printf("ERROR: Could not update avatar URL in DB: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	response := map[string]interface{}{
		"success": true,
		"data": map[string]string{
			"avatarUrl": avatarURL,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) listStudents(w http.ResponseWriter, r *http.Request) {
	// Authorization: Check if the user is a tutor
	role, _ := r.Context().Value(UserRoleKey).(string)
	if role != "tutor" {
		http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}

	students, pagination, err := h.userUseCase.ListStudents(r.Context(), page, limit)
	if err != nil {
		log.Printf("ERROR: failed to list students: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"students":   students,
			"pagination": pagination,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
