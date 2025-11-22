package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type StudyCycleHandler struct {
	svc      service.StudyCycleService
	validate *validator.Validate
}

func NewStudyCycleHandler(svc service.StudyCycleService) *StudyCycleHandler {
	return &StudyCycleHandler{svc: svc, validate: validator.New()}
}

type CreateStudyCycleRequest struct {
	Name        string `json:"name" validate:"required,min=2"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// CreateStudyCycle godoc
// @Summary Create a new study cycle
// @Tags study_cycles
// @Accept json
// @Produce json
// @Param input body CreateStudyCycleRequest true "Study cycle info"
// @Success 201 {object} database.StudyCycle
// @Router /study-cycles [post]
func (h *StudyCycleHandler) CreateStudyCycle(w http.ResponseWriter, r *http.Request) {
	var req CreateStudyCycleRequest
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

	cycle, err := h.svc.CreateStudyCycle(r.Context(), req.Name, req.Description, req.IsActive)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, cycle)
}

// GetActiveStudyCycle godoc
// @Summary Get the active study cycle
// @Tags study_cycles
// @Produce json
// @Success 200 {object} database.StudyCycle
// @Router /study-cycles/active [get]
func (h *StudyCycleHandler) GetActiveStudyCycle(w http.ResponseWriter, r *http.Request) {
	cycle, err := h.svc.GetActiveStudyCycle(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "No active study cycle found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, cycle)
}

func (h *StudyCycleHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *StudyCycleHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}
