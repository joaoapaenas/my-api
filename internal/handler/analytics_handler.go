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
// @Success 200 {array} handler.TimeReportResponse
// @Router /analytics/time-report [get]
func (h *AnalyticsHandler) GetTimeReport(w http.ResponseWriter, r *http.Request) {
	startDateFrom := r.URL.Query().Get("start_date_from")
	startDateTo := r.URL.Query().Get("start_date_to")

	report, err := h.svc.GetTimeReport(r.Context(), startDateFrom, startDateTo)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Map DB Result to JSON Response (DTO)
	response := make([]TimeReportResponse, len(report))
	for i, row := range report {
		response[i] = TimeReportResponse{
			SubjectID:     row.SubjectID,
			SubjectName:   row.SubjectName,
			ColorHex:      row.ColorHex.String, // Extract string from sql.NullString
			SessionsCount: int(row.SessionsCount),
			TotalHoursNet: row.TotalHoursNet,
		}
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetGlobalAccuracy godoc
// @Summary Get global accuracy by subject
// @Tags analytics
// @Produce json
// @Success 200 {array} handler.AccuracyReportResponse
// @Router /analytics/accuracy [get]
func (h *AnalyticsHandler) GetGlobalAccuracy(w http.ResponseWriter, r *http.Request) {
	report, err := h.svc.GetGlobalAccuracy(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Map DB Result to JSON Response
	response := make([]AccuracyReportResponse, len(report))
	for i, row := range report {
		response[i] = AccuracyReportResponse{
			SubjectID:          row.SubjectID,
			SubjectName:        row.SubjectName,
			ColorHex:           row.ColorHex.String,
			TotalQuestions:     int(row.TotalQuestions.Float64), // Handle sql.NullFloat64
			TotalCorrect:       int(row.TotalCorrect.Float64),
			AccuracyPercentage: row.AccuracyPercentage,
		}
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetWeakPoints godoc
// @Summary Get weak points (accuracy by topic) for a subject
// @Tags analytics
// @Produce json
// @Param subject_id path string true "Subject ID"
// @Success 200 {array} handler.TopicAccuracyResponse
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

	// Map DB Result to JSON Response
	response := make([]TopicAccuracyResponse, len(report))
	for i, row := range report {
		response[i] = TopicAccuracyResponse{
			TopicID:            row.TopicID,
			TopicName:          row.TopicName,
			TotalQuestions:     int(row.TotalQuestions.Float64),
			TotalCorrect:       int(row.TotalCorrect.Float64),
			AccuracyPercentage: row.AccuracyPercentage,
		}
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetHeatmap godoc
// @Summary Get study activity heatmap
// @Tags analytics
// @Produce json
// @Param days query int false "Number of days (default 30)"
// @Success 200 {array} handler.HeatmapDayResponse
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

	// Map DB Result to JSON Response
	response := make([]HeatmapDayResponse, len(heatmap))
	for i, row := range heatmap {
		// Handle interface{} types returned by SQLite driver for calculated fields
		dateStr, _ := row.StudyDate.(string)

		var totalSec int
		// TotalSeconds might come back as int64 or float64 depending on the driver/OS
		switch v := row.TotalSeconds.(type) {
		case int64:
			totalSec = int(v)
		case float64:
			totalSec = int(v)
		default:
			totalSec = 0
		}

		response[i] = HeatmapDayResponse{
			StudyDate:     dateStr,
			SessionsCount: int(row.SessionsCount),
			TotalSeconds:  totalSec,
		}
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *AnalyticsHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *AnalyticsHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}
