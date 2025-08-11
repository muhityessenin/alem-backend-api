package httpapi

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"

	"tutor/internal/domain"
	"tutor/internal/usecase"

	"github.com/gorilla/mux"
)

type ctxKey string

const (
	userIDKey ctxKey = "userID"
)

type TutorHandler struct {
	tutorUC usecase.TutorUseCase
	tokenUC usecase.TokenUseCase
}

func NewTutorHandler(t usecase.TutorUseCase, tok usecase.TokenUseCase) *TutorHandler {
	return &TutorHandler{tutorUC: t, tokenUC: tok}
}

func (h *TutorHandler) RegisterRoutes(r *mux.Router) {
	// public
	r.HandleFunc("/v1/tutors", h.listTutors).Methods("GET")
	r.HandleFunc("/v1/tutors/{id}", h.tutorDetails).Methods("GET")

	// protected wizard
	pr := r.PathPrefix("/v1/tutors/profile").Subrouter()
	pr.Use(h.jwt())
	pr.HandleFunc("/about", h.upsertAbout).Methods("PUT")
	pr.HandleFunc("/availability", h.replaceAvailability).Methods("PUT")
	pr.HandleFunc("/education", h.upsertEducation).Methods("PUT")
	pr.HandleFunc("/subjects", h.replaceSubjects).Methods("PUT")
	pr.HandleFunc("/video", h.setVideo).Methods("PUT")
	pr.HandleFunc("/complete", h.complete).Methods("POST")
}

// ---------- middleware ----------
func (h *TutorHandler) jwt() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hdr := r.Header.Get("Authorization")
			if hdr == "" {
				writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing Authorization")
				return
			}
			parts := strings.Split(hdr, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid Authorization")
				return
			}
			claims, err := h.tokenUC.ParseToken(r.Context(), parts[1])
			if err != nil {
				writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ---------- DTOs ----------
type aboutDTO struct {
	FirstName string                 `json:"firstName"`
	LastName  string                 `json:"lastName"`
	Phone     string                 `json:"phone"`
	Gender    string                 `json:"gender"`
	AvatarURL string                 `json:"avatarUrl"` // уже загруженный URL/ключ
	Languages []domain.TutorLanguage `json:"languages"`
}
type availabilityDTO struct {
	Timezone string                   `json:"timezone"`
	Days     []domain.AvailabilityDay `json:"days"`
}
type educationDTO struct {
	Description  string                 `json:"description"`
	Education    []domain.Education     `json:"education"`
	Certificates []domain.Certification `json:"certificates"`
}
type subjectsDTO struct {
	Items      []domain.TutorSubjectDTO `json:"items"`
	RegularMap map[string]int64         `json:"regularPrices"` // slug -> minor
	TrialMap   map[string]int64         `json:"trialPrices"`
}
type videoDTO struct {
	VideoURL string `json:"videoUrl"`
}

// ---------- handlers (wizard) ----------
func (h *TutorHandler) upsertAbout(w http.ResponseWriter, r *http.Request) {
	var req aboutDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "bad json")
		return
	}
	uid := r.Context().Value(userIDKey).(string)
	if err := h.tutorUC.UpsertAbout(r.Context(), uid, req.FirstName, req.LastName, req.Phone, req.Gender, req.AvatarURL, req.Languages); err != nil {
		writeErr(w, http.StatusInternalServerError, "UPsertAbout_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": true})
}

func (h *TutorHandler) replaceAvailability(w http.ResponseWriter, r *http.Request) {
	var req availabilityDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "bad json")
		return
	}
	if req.Timezone == "" {
		req.Timezone = "Asia/Almaty"
	}
	uid := r.Context().Value(userIDKey).(string)
	if err := h.tutorUC.ReplaceAvailability(r.Context(), uid, req.Timezone, req.Days); err != nil {
		writeErr(w, http.StatusInternalServerError, "AVAILABILITY_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": true})
}

func (h *TutorHandler) upsertEducation(w http.ResponseWriter, r *http.Request) {
	var req educationDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "bad json")
		return
	}
	uid := r.Context().Value(userIDKey).(string)
	if err := h.tutorUC.UpsertEducation(r.Context(), uid, req.Description, req.Education, req.Certificates); err != nil {
		writeErr(w, http.StatusInternalServerError, "EDU_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": true})
}

func (h *TutorHandler) replaceSubjects(w http.ResponseWriter, r *http.Request) {
	var req subjectsDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "bad json")
		return
	}
	uid := r.Context().Value(userIDKey).(string)
	if err := h.tutorUC.ReplaceSubjects(r.Context(), uid, req.Items, req.RegularMap, req.TrialMap); err != nil {
		writeErr(w, http.StatusInternalServerError, "SUBJECTS_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": true})
}

func (h *TutorHandler) setVideo(w http.ResponseWriter, r *http.Request) {
	var req videoDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "bad json")
		return
	}
	uid := r.Context().Value(userIDKey).(string)
	if err := h.tutorUC.SetVideo(r.Context(), uid, req.VideoURL); err != nil {
		writeErr(w, http.StatusInternalServerError, "VIDEO_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": true})
}

func (h *TutorHandler) complete(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(userIDKey).(string)
	if err := h.tutorUC.Complete(r.Context(), uid); err != nil {
		writeErr(w, http.StatusInternalServerError, "COMPLETE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": true})
}

// ---------- public list/details ----------
func (h *TutorHandler) listTutors(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 {
		limit = 20
	}

	filters := map[string]string{
		"search": q.Get("search"),
	}
	list, p, err := h.tutorUC.List(r.Context(), filters, page, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{
		"tutors": list, "pagination": p,
	}})
}

func (h *TutorHandler) tutorDetails(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	obj, err := h.tutorUC.GetDetails(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", "tutor not found")
		return
	}
	// клиенту может быть полезен IP/UA (пример)
	_ = net.IP{}.String
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": obj})
}

// ---------- helpers ----------
func writeErr(w http.ResponseWriter, code int, errCode, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error": map[string]string{
			"code": errCode, "message": msg,
		},
	})
}
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
