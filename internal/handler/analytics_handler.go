package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/joaoapaenas/my-api/internal/service"
)

type AnalyticsHandler struct {
	svc service.AnalyticsService
}

func NewAnalyticsHandler(svc service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

// GetTimeReportBySubject godoc
// @Summary Get time tracking report by subject
// @Tags analytics
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} handler.TimeReportResponse
// @Router /analytics/time-by-subject [get]
func (h *AnalyticsHandler) GetTimeReportBySubject(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	report, err := h.svc.GetTimeReportBySubject(r.Context(), startDate, endDate)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, report)
}

// GetAccuracyBySubject godoc
// @Summary Get accuracy report by subject
// @Tags analytics
// @Produce json
// @Success 200 {array} handler.AccuracyReportResponse
// @Router /analytics/accuracy-by-subject [get]
func (h *AnalyticsHandler) GetAccuracyBySubject(w http.ResponseWriter, r *http.Request) {
	report, err := h.svc.GetAccuracyBySubject(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, report)
}

// GetAccuracyByTopic godoc
// @Summary Get accuracy report by topic for a subject
// @Tags analytics
// @Produce json
// @Param subject_id path string true "Subject ID"
// @Success 200 {array} handler.TopicAccuracyResponse
// @Router /analytics/accuracy-by-topic/{subject_id} [get]
func (h *AnalyticsHandler) GetAccuracyByTopic(w http.ResponseWriter, r *http.Request) {
	subjectID := chi.URLParam(r, "subject_id")
	if subjectID == "" {
		h.respondWithError(w, http.StatusBadRequest, "Subject ID is required")
		return
	}

	report, err := h.svc.GetAccuracyByTopic(r.Context(), subjectID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, report)
}

// GetActivityHeatmap godoc
// @Summary Get activity heatmap data
// @Tags analytics
// @Produce json
// @Param days query int false "Number of days" default(30)
// @Success 200 {array} handler.HeatmapDayResponse
// @Router /analytics/heatmap [get]
func (h *AnalyticsHandler) GetActivityHeatmap(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 30 // default
	if daysStr != "" {
		if parsed, err := strconv.Atoi(daysStr); err == nil && parsed > 0 {
			days = parsed
		}
	}

	heatmap, err := h.svc.GetActivityHeatmap(r.Context(), days)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, heatmap)
}

func (h *AnalyticsHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *AnalyticsHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}
