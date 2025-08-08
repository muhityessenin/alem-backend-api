package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"tutor/internal/domain"

	"github.com/gorilla/mux"
	"tutor/internal/usecase"
)

type CtxKey string

const (
	UserIDKey   CtxKey = "userID"
	UserRoleKey CtxKey = "userRole"
)

type TutorHandler struct {
	useCase       usecase.TutorUseCase
	tokenUseCase  usecase.TokenUseCase
	reviewUseCase usecase.ReviewUseCase
}

func NewTutorHandler(tuc usecase.TutorUseCase, ruc usecase.ReviewUseCase, tokuc usecase.TokenUseCase) *TutorHandler {
	return &TutorHandler{useCase: tuc, reviewUseCase: ruc, tokenUseCase: tokuc}
}

func (h *TutorHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/tutors", h.listTutors).Methods("GET")
	router.HandleFunc("/api/tutors/{id}", h.getTutorDetails).Methods("GET")
	router.HandleFunc("/api/tutors/{id}/reviews", h.listTutorReviews).Methods("GET")

	protected := router.PathPrefix("/api/tutors").Subrouter()
	protected.Use(h.JWTMiddleware)
	protected.HandleFunc("/profile", h.updateTutorProfile).Methods("PUT")
	protected.HandleFunc("/reviews", h.createReview).Methods("POST") // <-- Our new route

}

func (h *TutorHandler) listTutors(w http.ResponseWriter, r *http.Request) {
	// --- Парсинг параметров запроса ---
	query := r.URL.Query()

	// Пагинация
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit < 1 {
		limit = 20 // Значение по умолчанию
	}

	// Фильтры
	filters := make(map[string]string)
	if search := query.Get("search"); search != "" {
		filters["search"] = search
	}
	// TODO: Добавить чтение других фильтров (subject, language, etc.)

	// --- Вызов бизнес-логики ---
	tutors, pagination, err := h.useCase.List(r.Context(), filters, page, limit)
	if err != nil {
		log.Printf("ERROR: Failed to list tutors: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// --- Формирование ответа ---
	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"tutors":     tutors,
			"pagination": pagination,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TutorHandler) getTutorDetails(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL path
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		log.Println("ERROR: Missing tutor ID in request")
		http.Error(w, "Tutor ID is required", http.StatusBadRequest)
		return
	}

	tutor, err := h.useCase.GetDetails(r.Context(), id)
	if err != nil {
		// TODO: Handle 'not found' error specifically
		log.Printf("ERROR: Failed to get tutor details for ID %s: %v", id, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    tutor,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// In alem-backend/tutor/internal/delivery/http/handler.go

// Replace your existing JWTMiddleware with this one.
func (h *TutorHandler) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Missing authorization header"}}`, http.StatusUnauthorized)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			http.Error(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Invalid authorization header format"}}`, http.StatusUnauthorized)
			return
		}

		tokenString := headerParts[1]
		claims, err := h.tokenUseCase.ParseToken(r.Context(), tokenString)
		if err != nil {
			log.Printf("WARN: Failed to parse token: %v", err)
			http.Error(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Invalid or expired token"}}`, http.StatusUnauthorized)
			return
		}

		// --- THIS IS THE FIX ---
		// We are explicitly using claims.UserID for the UserIDKey.
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		// We are not using the role here, but if we needed it, it would have its own key.
		// ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Add this new handler function inside handler.go
func (h *TutorHandler) updateTutorProfile(w http.ResponseWriter, r *http.Request) {
	// Get tutor ID from the token to ensure they only update their own profile
	tutorID, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}

	var tutorDetails domain.Tutor
	if err := json.NewDecoder(r.Body).Decode(&tutorDetails); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Important: Enforce that the ID from the token is used, not any ID from the request body
	tutorDetails.ID = tutorID

	updatedTutor, err := h.useCase.UpdateProfile(r.Context(), &tutorDetails)
	if err != nil {
		log.Printf("ERROR: Failed to update tutor profile for ID %s: %v", tutorID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    updatedTutor,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TutorHandler) listTutorReviews(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}

	reviews, pagination, stats, err := h.reviewUseCase.ListByTutor(r.Context(), id, page, limit)
	if err != nil {
		log.Printf("ERROR: Failed to list reviews for tutor %s: %v", id, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"reviews":            reviews,
			"pagination":         pagination,
			"averageRating":      stats.AverageRating,
			"ratingDistribution": stats.RatingDistribution,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TutorHandler) createReview(w http.ResponseWriter, r *http.Request) {
	// Get student ID from the token to ensure they are creating the review for themselves
	studentID, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}

	var req domain.Review
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Security: Enforce the student ID from the token
	req.StudentId = studentID

	createdReview, err := h.reviewUseCase.Create(r.Context(), &req)
	if err != nil {
		// This could be a "Conflict" error if a review for this booking already exists
		log.Printf("ERROR: Failed to create review for student %s: %v", studentID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    createdReview,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // Use 201 Created for new resources
	json.NewEncoder(w).Encode(response)
}

func (h *TutorHandler) listPendingTutors(w http.ResponseWriter, r *http.Request) {
	// Parse pagination from query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10 // Default limit
	}

	// Call the use case to get the data
	tutors, pagination, err := h.useCase.ListPending(r.Context(), page, limit)
	if err != nil {
		log.Printf("ERROR: Failed to list pending tutors: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Format and send the response
	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"tutors":     tutors,
			"pagination": pagination,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
