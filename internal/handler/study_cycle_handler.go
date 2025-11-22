package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
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

type UpdateStudyCycleRequest struct {
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

// GetStudyCycle godoc
// @Summary Get a study cycle by ID
// @Tags study_cycles
// @Produce json
// @Param id path string true "Cycle ID"
// @Success 200 {object} database.StudyCycle
// @Router /study-cycles/{id} [get]
func (h *StudyCycleHandler) GetStudyCycle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Cycle ID is required")
		return
	}

	cycle, err := h.svc.GetStudyCycle(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Study cycle not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, cycle)
}

// UpdateStudyCycle godoc
// @Summary Update a study cycle
// @Tags study_cycles
// @Accept json
// @Produce json
// @Param id path string true "Cycle ID"
// @Param input body UpdateStudyCycleRequest true "Study cycle info"
// @Success 200 {string} string "OK"
// @Router /study-cycles/{id} [put]
func (h *StudyCycleHandler) UpdateStudyCycle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Cycle ID is required")
		return
	}

	var req UpdateStudyCycleRequest
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

	err := h.svc.UpdateStudyCycle(r.Context(), id, req.Name, req.Description, req.IsActive)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Study cycle updated successfully"})
}

// DeleteStudyCycle godoc
// @Summary Delete a study cycle
// @Tags study_cycles
// @Param id path string true "Cycle ID"
// @Success 204
// @Router /study-cycles/{id} [delete]
func (h *StudyCycleHandler) DeleteStudyCycle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Cycle ID is required")
		return
	}

	err := h.svc.DeleteStudyCycle(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *StudyCycleHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *StudyCycleHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}
