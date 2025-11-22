package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type StudySessionHandler struct {
	svc      service.StudySessionService
	validate *validator.Validate
}

func NewStudySessionHandler(svc service.StudySessionService) *StudySessionHandler {
	return &StudySessionHandler{svc: svc, validate: validator.New()}
}

type CreateStudySessionRequest struct {
	SubjectID   string `json:"subject_id" validate:"required"`
	CycleItemID string `json:"cycle_item_id"`
	StartedAt   string `json:"started_at" validate:"required"`
}

type UpdateSessionDurationRequest struct {
	FinishedAt           string `json:"finished_at"`
	GrossDurationSeconds int    `json:"gross_duration_seconds"`
	NetDurationSeconds   int    `json:"net_duration_seconds"`
	Notes                string `json:"notes"`
}

// CreateStudySession godoc
// @Summary Create a new study session
// @Tags study_sessions
// @Accept json
// @Produce json
// @Param input body CreateStudySessionRequest true "Study session info"
// @Success 201 {object} database.StudySession
// @Router /study-sessions [post]
func (h *StudySessionHandler) CreateStudySession(w http.ResponseWriter, r *http.Request) {
	var req CreateStudySessionRequest
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

	session, err := h.svc.CreateStudySession(r.Context(), req.SubjectID, req.CycleItemID, req.StartedAt)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, session)
}

// UpdateSessionDuration godoc
// @Summary Update study session duration
// @Tags study_sessions
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Param input body UpdateSessionDurationRequest true "Session duration info"
// @Success 200 {string} string "OK"
// @Router /study-sessions/{id} [put]
func (h *StudySessionHandler) UpdateSessionDuration(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		h.respondWithError(w, http.StatusBadRequest, "Session ID is required")
		return
	}

	var req UpdateSessionDurationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.svc.UpdateSessionDuration(r.Context(), sessionID, req.FinishedAt, req.GrossDurationSeconds, req.NetDurationSeconds, req.Notes)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Session updated successfully"})
}

func (h *StudySessionHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *StudySessionHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}
