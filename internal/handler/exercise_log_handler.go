package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type ExerciseLogHandler struct {
	svc      service.ExerciseLogService
	validate *validator.Validate
}

func NewExerciseLogHandler(svc service.ExerciseLogService) *ExerciseLogHandler {
	return &ExerciseLogHandler{svc: svc, validate: validator.New()}
}

type CreateExerciseLogRequest struct {
	SessionID      string `json:"session_id"`
	SubjectID      string `json:"subject_id" validate:"required"`
	TopicID        string `json:"topic_id"`
	QuestionsCount int    `json:"questions_count" validate:"required,min=0"`
	CorrectCount   int    `json:"correct_count" validate:"required,min=0"`
}

// CreateExerciseLog godoc
// @Summary Create a new exercise log
// @Tags exercise_logs
// @Accept json
// @Produce json
// @Param input body CreateExerciseLogRequest true "Exercise log info"
// @Success 201 {object} database.ExerciseLog
// @Router /exercise-logs [post]
func (h *ExerciseLogHandler) CreateExerciseLog(w http.ResponseWriter, r *http.Request) {
	var req CreateExerciseLogRequest
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

	// Validate that correct_count <= questions_count
	if req.CorrectCount > req.QuestionsCount {
		h.respondWithError(w, http.StatusBadRequest, "Correct count cannot exceed questions count")
		return
	}

	log, err := h.svc.CreateExerciseLog(r.Context(), req.SessionID, req.SubjectID, req.TopicID, req.QuestionsCount, req.CorrectCount)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, log)
}

func (h *ExerciseLogHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *ExerciseLogHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}
