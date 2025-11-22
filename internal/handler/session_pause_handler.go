package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type SessionPauseHandler struct {
	svc      service.SessionPauseService
	validate *validator.Validate
}

func NewSessionPauseHandler(svc service.SessionPauseService) *SessionPauseHandler {
	return &SessionPauseHandler{svc: svc, validate: validator.New()}
}

type CreateSessionPauseRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	StartedAt string `json:"started_at" validate:"required"`
}

type EndSessionPauseRequest struct {
	EndedAt string `json:"ended_at" validate:"required"`
}

// CreateSessionPause godoc
// @Summary Create a new session pause
// @Tags session_pauses
// @Accept json
// @Produce json
// @Param input body CreateSessionPauseRequest true "Session pause info"
// @Success 201 {object} handler.SessionPauseResponse
// @Router /session-pauses [post]
func (h *SessionPauseHandler) CreateSessionPause(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionPauseRequest
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

	pause, err := h.svc.CreateSessionPause(r.Context(), req.SessionID, req.StartedAt)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, pause)
}

// GetSessionPause godoc
// @Summary Get a session pause by ID
// @Tags session_pauses
// @Produce json
// @Param id path string true "Pause ID"
// @Success 200 {object} handler.SessionPauseResponse
// @Router /session-pauses/{id} [get]
func (h *SessionPauseHandler) GetSessionPause(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Pause ID is required")
		return
	}

	pause, err := h.svc.GetSessionPause(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Session pause not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, pause)
}

// EndSessionPause godoc
// @Summary End a session pause
// @Tags session_pauses
// @Accept json
// @Produce json
// @Param id path string true "Pause ID"
// @Param input body EndSessionPauseRequest true "End pause info"
// @Success 200 {string} string "OK"
// @Router /session-pauses/{id}/end [put]
func (h *SessionPauseHandler) EndSessionPause(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Pause ID is required")
		return
	}

	var req EndSessionPauseRequest
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

	err := h.svc.EndSessionPause(r.Context(), id, req.EndedAt)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Pause ended successfully"})
}

// DeleteSessionPause godoc
// @Summary Delete a session pause
// @Tags session_pauses
// @Param id path string true "Pause ID"
// @Success 204
// @Router /session-pauses/{id} [delete]
func (h *SessionPauseHandler) DeleteSessionPause(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Pause ID is required")
		return
	}

	err := h.svc.DeleteSessionPause(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionPauseHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *SessionPauseHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}
