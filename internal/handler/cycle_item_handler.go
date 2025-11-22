package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type CycleItemHandler struct {
	svc      service.CycleItemService
	validate *validator.Validate
}

func NewCycleItemHandler(svc service.CycleItemService) *CycleItemHandler {
	return &CycleItemHandler{svc: svc, validate: validator.New()}
}

type CreateCycleItemRequest struct {
	SubjectID              string `json:"subject_id" validate:"required"`
	OrderIndex             int    `json:"order_index" validate:"required,min=1"`
	PlannedDurationMinutes int    `json:"planned_duration_minutes" validate:"omitempty,min=1"`
}

type UpdateCycleItemRequest struct {
	SubjectID              string `json:"subject_id" validate:"required"`
	OrderIndex             int    `json:"order_index" validate:"required,min=1"`
	PlannedDurationMinutes int    `json:"planned_duration_minutes" validate:"omitempty,min=1"`
}

// CreateCycleItem godoc
// @Summary Create a new cycle item
// @Tags cycle_items
// @Accept json
// @Produce json
// @Param id path string true "Cycle ID"
// @Param input body CreateCycleItemRequest true "Cycle item info"
// @Success 201 {object} database.CycleItem
// @Router /study-cycles/{id}/items [post]
func (h *CycleItemHandler) CreateCycleItem(w http.ResponseWriter, r *http.Request) {
	cycleID := chi.URLParam(r, "id")
	if cycleID == "" {
		h.respondWithError(w, http.StatusBadRequest, "Cycle ID is required")
		return
	}

	var req CreateCycleItemRequest
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

	item, err := h.svc.CreateCycleItem(r.Context(), cycleID, req.SubjectID, req.OrderIndex, req.PlannedDurationMinutes)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, item)
}

// ListCycleItems godoc
// @Summary List all items for a cycle
// @Tags cycle_items
// @Produce json
// @Param id path string true "Cycle ID"
// @Success 200 {array} database.CycleItem
// @Router /study-cycles/{id}/items [get]
func (h *CycleItemHandler) ListCycleItems(w http.ResponseWriter, r *http.Request) {
	cycleID := chi.URLParam(r, "id")
	if cycleID == "" {
		h.respondWithError(w, http.StatusBadRequest, "Cycle ID is required")
		return
	}

	items, err := h.svc.ListCycleItems(r.Context(), cycleID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, items)
}

// GetCycleItem godoc
// @Summary Get a cycle item by ID
// @Tags cycle_items
// @Produce json
// @Param id path string true "Item ID"
// @Success 200 {object} database.CycleItem
// @Router /cycle-items/{id} [get]
func (h *CycleItemHandler) GetCycleItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Item ID is required")
		return
	}

	item, err := h.svc.GetCycleItem(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Cycle item not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, item)
}

// UpdateCycleItem godoc
// @Summary Update a cycle item
// @Tags cycle_items
// @Accept json
// @Produce json
// @Param id path string true "Item ID"
// @Param input body UpdateCycleItemRequest true "Cycle item info"
// @Success 200 {string} string "OK"
// @Router /cycle-items/{id} [put]
func (h *CycleItemHandler) UpdateCycleItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Item ID is required")
		return
	}

	var req UpdateCycleItemRequest
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

	err := h.svc.UpdateCycleItem(r.Context(), id, req.SubjectID, req.OrderIndex, req.PlannedDurationMinutes)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Cycle item updated successfully"})
}

// DeleteCycleItem godoc
// @Summary Delete a cycle item
// @Tags cycle_items
// @Param id path string true "Item ID"
// @Success 204
// @Router /cycle-items/{id} [delete]
func (h *CycleItemHandler) DeleteCycleItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Item ID is required")
		return
	}

	err := h.svc.DeleteCycleItem(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CycleItemHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *CycleItemHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}
