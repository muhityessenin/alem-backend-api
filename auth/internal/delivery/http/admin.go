// auth/internal/delivery/http/admin.go
package http

import (
	"auth/internal/domain"
	_ "auth/internal/repository"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// ------------ middleware (admin only) ------------
func (h *AuthHandler) adminOnly() mux.MiddlewareFunc {
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
			claims, err := h.useCase.ParseToken(r.Context(), parts[1])
			if err != nil {
				writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
				return
			}
			if claims.Role != domain.AdminRole {
				writeErr(w, http.StatusForbidden, "FORBIDDEN", "admin only")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ------------ роутер (добавка к твоим RegisterRoutes) ------------
// auth/internal/delivery/http/admin.go
func (h *AuthHandler) RegisterAdminRoutes(r *mux.Router) {
	// Публичные GET под тем же префиксом
	public := r.PathPrefix("/api/v1/admin").Subrouter()

	// languages (GET — публично)
	public.HandleFunc("/languages", h.adminListLanguages).Methods(http.MethodGet)

	// subjects (GET — публично)
	public.HandleFunc("/subjects", h.adminListSubjects).Methods(http.MethodGet)

	// directions / subdirections (GET — публично)
	public.HandleFunc("/directions", h.adminListDirections).Methods(http.MethodGet)
	public.HandleFunc("/subdirections", h.adminListSubdirections).Methods(http.MethodGet)

	// tutor <-> taxonomy bindings (GET — публично)
	public.HandleFunc("/tutors/{tutorID}/subjects", h.adminListTutorSubjects).Methods(http.MethodGet)
	public.HandleFunc("/tutors/{tutorID}/languages", h.adminListTutorLanguages).Methods(http.MethodGet)
	public.HandleFunc("/tutors/{tutorID}/subdirections", h.adminListTutorSubdirections).Methods(http.MethodGet)

	// Защищённые POST — только для админа
	protected := r.PathPrefix("/api/v1/admin").Subrouter()
	protected.Use(h.adminOnly())

	// languages (POST — только админ)
	protected.HandleFunc("/languages", h.adminCreateLanguage).Methods(http.MethodPost, http.MethodOptions)

	// subjects (POST — только админ)
	protected.HandleFunc("/subjects", h.adminCreateSubject).Methods(http.MethodPost, http.MethodOptions)

	// directions / subdirections (POST — только админ)
	protected.HandleFunc("/directions", h.adminCreateDirection).Methods(http.MethodPost, http.MethodOptions)
	protected.HandleFunc("/subdirections", h.adminCreateSubdirection).Methods(http.MethodPost, http.MethodOptions)

	// tutor <-> taxonomy bindings (POST — только админ)
	protected.HandleFunc("/tutors/{tutorID}/subjects", h.adminUpsertTutorSubject).Methods(http.MethodPost, http.MethodOptions)
	protected.HandleFunc("/tutors/{tutorID}/languages", h.adminUpsertTutorLanguage).Methods(http.MethodPost, http.MethodOptions)
	protected.HandleFunc("/tutors/{tutorID}/subdirections", h.adminUpsertTutorSubdirection).Methods(http.MethodPost, http.MethodOptions)
}

// ------------ DTOs ------------
type kvMap = map[string]string

type languageDTO struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type subjectDTO struct {
	Slug string         `json:"slug"`
	Name map[string]any `json:"name"` // jsonb ({"ru":"...", "en":"..."})
}

type directionDTO struct {
	Slug string         `json:"slug"`
	Name map[string]any `json:"name"`
}

type subdirectionDTO struct {
	DirectionID   string         `json:"directionId,omitempty"`
	DirectionSlug string         `json:"directionSlug,omitempty"`
	Slug          string         `json:"slug"`
	Name          map[string]any `json:"name"`
}

type tutorSubjectDTO struct {
	SubjectID   string `json:"subjectId,omitempty"`
	SubjectSlug string `json:"subjectSlug,omitempty"`
	Level       string `json:"level,omitempty"`
	PriceMinor  int64  `json:"priceMinor,omitempty"`
	Currency    string `json:"currency,omitempty"` // default KZT
}

type tutorLanguageDTO struct {
	Code        string `json:"code"`
	Proficiency string `json:"proficiency"` // A1..C2/native
}

type tutorSubdirectionDTO struct {
	SubdirectionID   string `json:"subdirectionId,omitempty"`
	SubdirectionSlug string `json:"subdirectionSlug,omitempty"`
	Level            string `json:"level,omitempty"`
	PriceMinor       int64  `json:"priceMinor,omitempty"`
	Currency         string `json:"currency,omitempty"`
}

// ------------ handlers: languages ------------
func (h *AuthHandler) adminCreateLanguage(w http.ResponseWriter, r *http.Request) {
	var req languageDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" || req.Name == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "code and name required")
		return
	}
	if err := h.adminUC.CreateLanguage(r.Context(), req.Code, req.Name); err != nil {
		writeErr(w, http.StatusConflict, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true})
}

func (h *AuthHandler) adminListLanguages(w http.ResponseWriter, r *http.Request) {
	items, err := h.adminUC.ListLanguages(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": items})
}

// ------------ handlers: subjects ------------
func (h *AuthHandler) adminCreateSubject(w http.ResponseWriter, r *http.Request) {
	var req subjectDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Slug == "" || len(req.Name) == 0 {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "slug and name required")
		return
	}
	if err := h.adminUC.CreateSubject(r.Context(), req.Slug, req.Name); err != nil {
		writeErr(w, http.StatusConflict, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true})
}

func (h *AuthHandler) adminListSubjects(w http.ResponseWriter, r *http.Request) {
	items, err := h.adminUC.ListSubjects(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": items})
}

// ------------ handlers: directions / subdirections ------------
func (h *AuthHandler) adminCreateDirection(w http.ResponseWriter, r *http.Request) {
	var req directionDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Slug == "" || len(req.Name) == 0 {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "slug and name required")
		return
	}
	if err := h.adminUC.CreateDirection(r.Context(), req.Slug, req.Name); err != nil {
		writeErr(w, http.StatusConflict, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true})
}

func (h *AuthHandler) adminListDirections(w http.ResponseWriter, r *http.Request) {
	items, err := h.adminUC.ListDirections(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": items})
}

func (h *AuthHandler) adminCreateSubdirection(w http.ResponseWriter, r *http.Request) {
	var req subdirectionDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Slug == "" || len(req.Name) == 0 {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "slug and name required")
		return
	}
	if err := h.adminUC.CreateSubdirection(r.Context(), req.DirectionID, req.DirectionSlug, req.Slug, req.Name); err != nil {
		writeErr(w, http.StatusConflict, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true})
}

func (h *AuthHandler) adminListSubdirections(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("direction")
	items, err := h.adminUC.ListSubdirections(r.Context(), dir)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": items})
}

// ------------ handlers: tutor bindings ------------
func (h *AuthHandler) adminUpsertTutorSubject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tutorID := vars["tutorID"]
	var req tutorSubjectDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "bad json")
		return
	}
	if err := h.adminUC.UpsertTutorSubject(r.Context(), tutorID, req.SubjectID, req.SubjectSlug, req.Level, req.PriceMinor, req.Currency); err != nil {
		writeErr(w, http.StatusConflict, "UPSERT_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *AuthHandler) adminListTutorSubjects(w http.ResponseWriter, r *http.Request) {
	tutorID := mux.Vars(r)["tutorID"]
	items, err := h.adminUC.ListTutorSubjects(r.Context(), tutorID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": items})
}

func (h *AuthHandler) adminUpsertTutorLanguage(w http.ResponseWriter, r *http.Request) {
	tutorID := mux.Vars(r)["tutorID"]
	var req tutorLanguageDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "bad json")
		return
	}
	if err := h.adminUC.UpsertTutorLanguage(r.Context(), tutorID, req.Code, req.Proficiency); err != nil {
		writeErr(w, http.StatusConflict, "UPSERT_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *AuthHandler) adminListTutorLanguages(w http.ResponseWriter, r *http.Request) {
	tutorID := mux.Vars(r)["tutorID"]
	items, err := h.adminUC.ListTutorLanguages(r.Context(), tutorID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": items})
}

func (h *AuthHandler) adminUpsertTutorSubdirection(w http.ResponseWriter, r *http.Request) {
	tutorID := mux.Vars(r)["tutorID"]
	var req tutorSubdirectionDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "bad json")
		return
	}
	if err := h.adminUC.UpsertTutorSubdirection(r.Context(), tutorID, req.SubdirectionID, req.SubdirectionSlug, req.Level, req.PriceMinor, req.Currency); err != nil {
		writeErr(w, http.StatusConflict, "UPSERT_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *AuthHandler) adminListTutorSubdirections(w http.ResponseWriter, r *http.Request) {
	tutorID := mux.Vars(r)["tutorID"]
	items, err := h.adminUC.ListTutorSubdirections(r.Context(), tutorID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": items})
}
