package service

import (
	"context"
	"fmt"

	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type AnalyticsService interface {
	GetTimeReport(ctx context.Context, startDateFrom, startDateTo string) ([]database.GetTimeReportBySubjectRow, error)
	GetGlobalAccuracy(ctx context.Context) ([]database.GetAccuracyBySubjectRow, error)
	GetWeakPoints(ctx context.Context, subjectID string) ([]database.GetAccuracyByTopicRow, error)
	GetHeatmap(ctx context.Context, daysCount int64) ([]database.GetActivityHeatmapRow, error)
}

type AnalyticsManager struct {
	repo repository.AnalyticsRepository
}

func NewAnalyticsManager(repo repository.AnalyticsRepository) *AnalyticsManager {
	return &AnalyticsManager{repo: repo}
}

func (s *AnalyticsManager) GetTimeReport(ctx context.Context, startDateFrom, startDateTo string) ([]database.GetTimeReportBySubjectRow, error) {
	return s.repo.GetTimeReportBySubject(ctx, database.GetTimeReportBySubjectParams{
		StartDateFrom: startDateFrom,
		StartDateTo:   startDateTo,
	})
}

func (s *AnalyticsManager) GetGlobalAccuracy(ctx context.Context) ([]database.GetAccuracyBySubjectRow, error) {
	return s.repo.GetAccuracyBySubject(ctx)
}

func (s *AnalyticsManager) GetWeakPoints(ctx context.Context, subjectID string) ([]database.GetAccuracyByTopicRow, error) {
	return s.repo.GetAccuracyByTopic(ctx, subjectID)
}

func (s *AnalyticsManager) GetHeatmap(ctx context.Context, daysCount int64) ([]database.GetActivityHeatmapRow, error) {
	// Default to 30 days if 0 or negative
	if daysCount <= 0 {
		daysCount = 30
	}
	return s.repo.GetActivityHeatmap(ctx, fmt.Sprintf("%d", daysCount))
}
