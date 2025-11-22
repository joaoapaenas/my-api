package service

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type AnalyticsService interface {
	GetTimeReportBySubject(ctx context.Context, startDate, endDate string) ([]database.GetTimeReportBySubjectRow, error)
	GetAccuracyBySubject(ctx context.Context) ([]database.GetAccuracyBySubjectRow, error)
	GetAccuracyByTopic(ctx context.Context, subjectID string) ([]database.GetAccuracyByTopicRow, error)
	GetActivityHeatmap(ctx context.Context, days int) ([]database.GetActivityHeatmapRow, error)
}

type AnalyticsManager struct {
	repo repository.AnalyticsRepository
}

func NewAnalyticsManager(repo repository.AnalyticsRepository) *AnalyticsManager {
	return &AnalyticsManager{repo: repo}
}

func (s *AnalyticsManager) GetTimeReportBySubject(ctx context.Context, startDate, endDate string) ([]database.GetTimeReportBySubjectRow, error) {
	// Prepare parameters for the query
	// Empty strings mean no filter
	return s.repo.GetTimeReportBySubject(ctx, database.GetTimeReportBySubjectParams{
		Column1: startDate,
		Column2: startDate,
		Column3: endDate,
		Column4: endDate,
	})
}

func (s *AnalyticsManager) GetAccuracyBySubject(ctx context.Context) ([]database.GetAccuracyBySubjectRow, error) {
	return s.repo.GetAccuracyBySubject(ctx)
}

func (s *AnalyticsManager) GetAccuracyByTopic(ctx context.Context, subjectID string) ([]database.GetAccuracyByTopicRow, error) {
	return s.repo.GetAccuracyByTopic(ctx, subjectID)
}

func (s *AnalyticsManager) GetActivityHeatmap(ctx context.Context, days int) ([]database.GetActivityHeatmapRow, error) {
	return s.repo.GetActivityHeatmap(ctx, int64(days))
}
