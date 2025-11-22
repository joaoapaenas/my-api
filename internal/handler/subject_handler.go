package handler

import (
	"encoding/json"
	"net/http"

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

// CreateSubject godoc
// @Summary Create a new subject
// @Tags subjects
// @Accept json
// @Produce json
// @Param input body CreateSubjectRequest true "Subject info"
// @Success 201 {object} database.Subject
// @Router /subjects [post]
func (h *SubjectHandler) CreateSubject(w http.ResponseWriter, r *http.Request) {
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

	subject, err := h.svc.CreateSubject(r.Context(), req.Name, req.ColorHex)
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
// @Success 200 {array} database.Subject
// @Router /subjects [get]
func (h *SubjectHandler) ListSubjects(w http.ResponseWriter, r *http.Request) {
	subjects, err := h.svc.ListSubjects(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, subjects)
}

func (h *SubjectHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *SubjectHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}
