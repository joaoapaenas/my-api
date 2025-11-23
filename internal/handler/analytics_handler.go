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

// GetTimeReport godoc
// @Summary Get net study time report by subject
// @Tags analytics
// @Produce json
// @Param start_date_from query string false "Start Date From (YYYY-MM-DD)"
// @Param start_date_to query string false "Start Date To (YYYY-MM-DD)"
// @Success 200 {array} database.GetTimeReportBySubjectRow
// @Router /analytics/time-report [get]
func (h *AnalyticsHandler) GetTimeReport(w http.ResponseWriter, r *http.Request) {
	startDateFrom := r.URL.Query().Get("start_date_from")
	startDateTo := r.URL.Query().Get("start_date_to")

	report, err := h.svc.GetTimeReport(r.Context(), startDateFrom, startDateTo)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, report)
}

// GetGlobalAccuracy godoc
// @Summary Get global accuracy by subject
// @Tags analytics
// @Produce json
// @Success 200 {array} database.GetAccuracyBySubjectRow
// @Router /analytics/accuracy [get]
func (h *AnalyticsHandler) GetGlobalAccuracy(w http.ResponseWriter, r *http.Request) {
	report, err := h.svc.GetGlobalAccuracy(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, report)
}

// GetWeakPoints godoc
// @Summary Get weak points (accuracy by topic) for a subject
// @Tags analytics
// @Produce json
// @Param subject_id path string true "Subject ID"
// @Success 200 {array} database.GetAccuracyByTopicRow
// @Router /analytics/weak-points/{subject_id} [get]
func (h *AnalyticsHandler) GetWeakPoints(w http.ResponseWriter, r *http.Request) {
	subjectID := chi.URLParam(r, "subject_id")
	if subjectID == "" {
		h.respondWithError(w, http.StatusBadRequest, "Subject ID is required")
		return
	}

	report, err := h.svc.GetWeakPoints(r.Context(), subjectID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, report)
}

// GetHeatmap godoc
// @Summary Get study activity heatmap
// @Tags analytics
// @Produce json
// @Param days query int false "Number of days (default 30)"
// @Success 200 {array} database.GetActivityHeatmapRow
// @Router /analytics/heatmap [get]
func (h *AnalyticsHandler) GetHeatmap(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	var days int64 = 30
	if daysStr != "" {
		if d, err := strconv.ParseInt(daysStr, 10, 64); err == nil {
			days = d
		}
	}

	heatmap, err := h.svc.GetHeatmap(r.Context(), days)
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
