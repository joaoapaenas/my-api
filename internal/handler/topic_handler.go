package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type TopicHandler struct {
	svc      service.TopicService
	validate *validator.Validate
}

func NewTopicHandler(svc service.TopicService) *TopicHandler {
	return &TopicHandler{svc: svc, validate: validator.New()}
}

type CreateTopicRequest struct {
	Name string `json:"name" validate:"required,min=2"`
}

type UpdateTopicRequest struct {
	Name string `json:"name" validate:"required,min=2"`
}

// CreateTopic godoc
// @Summary Create a new topic for a subject
// @Tags topics
// @Accept json
// @Produce json
// @Param id path string true "Subject ID"
// @Param input body CreateTopicRequest true "Topic info"
// @Success 201 {object} database.Topic
// @Router /subjects/{id}/topics [post]
func (h *TopicHandler) CreateTopic(w http.ResponseWriter, r *http.Request) {
	subjectID := chi.URLParam(r, "id")
	if subjectID == "" {
		h.respondWithError(w, http.StatusBadRequest, "Subject ID is required")
		return
	}

	var req CreateTopicRequest
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

	topic, err := h.svc.CreateTopic(r.Context(), subjectID, req.Name)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, topic)
}

// ListTopics godoc
// @Summary List all topics for a subject
// @Tags topics
// @Produce json
// @Param id path string true "Subject ID"
// @Success 200 {array} database.Topic
// @Router /subjects/{id}/topics [get]
func (h *TopicHandler) ListTopics(w http.ResponseWriter, r *http.Request) {
	subjectID := chi.URLParam(r, "id")
	if subjectID == "" {
		h.respondWithError(w, http.StatusBadRequest, "Subject ID is required")
		return
	}

	topics, err := h.svc.ListTopicsBySubject(r.Context(), subjectID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, topics)
}

// GetTopic godoc
// @Summary Get a topic by ID
// @Tags topics
// @Produce json
// @Param id path string true "Topic ID"
// @Success 200 {object} database.Topic
// @Router /topics/{id} [get]
func (h *TopicHandler) GetTopic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Topic ID is required")
		return
	}

	topic, err := h.svc.GetTopic(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Topic not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, topic)
}

// UpdateTopic godoc
// @Summary Update a topic
// @Tags topics
// @Accept json
// @Produce json
// @Param id path string true "Topic ID"
// @Param input body UpdateTopicRequest true "Topic info"
// @Success 200 {string} string "OK"
// @Router /topics/{id} [put]
func (h *TopicHandler) UpdateTopic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Topic ID is required")
		return
	}

	var req UpdateTopicRequest
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

	err := h.svc.UpdateTopic(r.Context(), id, req.Name)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Topic updated successfully"})
}

// DeleteTopic godoc
// @Summary Delete a topic
// @Tags topics
// @Param id path string true "Topic ID"
// @Success 204
// @Router /topics/{id} [delete]
func (h *TopicHandler) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Topic ID is required")
		return
	}

	err := h.svc.DeleteTopic(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TopicHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *TopicHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}
