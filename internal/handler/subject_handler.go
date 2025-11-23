package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type SubjectHandler struct {
	svc      service.SubjectService
	validate *validator.Validate
}

func NewSubjectHandler(svc service.SubjectService) *SubjectHandler {
	return &SubjectHandler{svc: svc, validate: validator.New()}
}

type CreateSubjectRequest struct {
	Name     string `json:"name" validate:"required,min=2"`
	ColorHex string `json:"color_hex" validate:"omitempty,hexcolor"`
}

type UpdateSubjectRequest struct {
	Name     string `json:"name" validate:"required,min=2"`
	ColorHex string `json:"color_hex" validate:"omitempty,hexcolor"`
}

// CreateSubject godoc
// @Summary Create a new subject
// @Tags subjects
// @Accept json
// @Produce json
// @Param input body CreateSubjectRequest true "Subject info"
// @Success 201 {object} handler.SubjectResponse
// @Router /subjects [post]
func (h *SubjectHandler) CreateSubject(w http.ResponseWriter, r *http.Request) {
	// 1. Extract UserID
	userID := r.Context().Value("userID")
	if userID == nil {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req CreateSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	// 2. Pass userID to Service
	subject, err := h.svc.CreateSubject(r.Context(), userID.(string), req.Name, req.ColorHex)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, subject)
}

// ListSubjects godoc
// @Summary List all subjects
// @Tags subjects
// @Produce json
// @Success 200 {array} handler.SubjectResponse
// @Router /subjects [get]
func (h *SubjectHandler) ListSubjects(w http.ResponseWriter, r *http.Request) {
	// 1. Extract UserID
	userID := r.Context().Value("userID")
	if userID == nil {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// 2. Pass userID to Service
	subjects, err := h.svc.ListSubjects(r.Context(), userID.(string))
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, subjects)
}

// GetSubject godoc
// @Summary Get a subject by ID
// @Tags subjects
// @Produce json
// @Param id path string true "Subject ID"
// @Success 200 {object} handler.SubjectResponse
// @Router /subjects/{id} [get]
func (h *SubjectHandler) GetSubject(w http.ResponseWriter, r *http.Request) {
	// 1. Extract UserID
	userID := r.Context().Value("userID")
	if userID == nil {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Subject ID is required")
		return
	}

	// 2. Pass userID to Service
	subject, err := h.svc.GetSubject(r.Context(), id, userID.(string))
	if err != nil {
		// If DB returns nothing because userID didn't match, it looks like a generic "Not Found", which is correct security.
		h.respondWithError(w, http.StatusNotFound, "Subject not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, subject)
}

// UpdateSubject godoc
// @Summary Update a subject
// @Tags subjects
// @Accept json
// @Produce json
// @Param id path string true "Subject ID"
// @Param input body UpdateSubjectRequest true "Subject info"
// @Success 200 {string} string "OK"
// @Router /subjects/{id} [put]
func (h *SubjectHandler) UpdateSubject(w http.ResponseWriter, r *http.Request) {
	// 1. Extract UserID
	userID := r.Context().Value("userID")
	if userID == nil {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Subject ID is required")
		return
	}

	var req UpdateSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	// 2. Pass userID to Service
	err := h.svc.UpdateSubject(r.Context(), id, userID.(string), req.Name, req.ColorHex)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Subject updated successfully"})
}

// DeleteSubject godoc
// @Summary Delete a subject
// @Tags subjects
// @Param id path string true "Subject ID"
// @Success 204
// @Router /subjects/{id} [delete]
func (h *SubjectHandler) DeleteSubject(w http.ResponseWriter, r *http.Request) {
	// 1. Extract UserID
	userID := r.Context().Value("userID")
	if userID == nil {
		h.respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Subject ID is required")
		return
	}

	// 2. Pass userID to Service
	err := h.svc.DeleteSubject(r.Context(), id, userID.(string))
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SubjectHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *SubjectHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}
