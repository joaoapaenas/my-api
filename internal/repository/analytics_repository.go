package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type AnalyticsRepository interface {
	GetTimeReportBySubject(ctx context.Context, arg database.GetTimeReportBySubjectParams) ([]database.GetTimeReportBySubjectRow, error)
	GetAccuracyBySubject(ctx context.Context) ([]database.GetAccuracyBySubjectRow, error)
	GetAccuracyByTopic(ctx context.Context, subjectID string) ([]database.GetAccuracyByTopicRow, error)
	GetActivityHeatmap(ctx context.Context, daysCount string) ([]database.GetActivityHeatmapRow, error)
}

type SQLAnalyticsRepository struct {
	q database.Querier
}

func NewSQLAnalyticsRepository(q database.Querier) *SQLAnalyticsRepository {
	return &SQLAnalyticsRepository{q: q}
}

func (r *SQLAnalyticsRepository) GetTimeReportBySubject(ctx context.Context, arg database.GetTimeReportBySubjectParams) ([]database.GetTimeReportBySubjectRow, error) {
	return r.q.GetTimeReportBySubject(ctx, arg)
}

func (r *SQLAnalyticsRepository) GetAccuracyBySubject(ctx context.Context) ([]database.GetAccuracyBySubjectRow, error) {
	return r.q.GetAccuracyBySubject(ctx)
}

func (r *SQLAnalyticsRepository) GetAccuracyByTopic(ctx context.Context, subjectID string) ([]database.GetAccuracyByTopicRow, error) {
	return r.q.GetAccuracyByTopic(ctx, subjectID)
}

func (r *SQLAnalyticsRepository) GetActivityHeatmap(ctx context.Context, daysCount string) ([]database.GetActivityHeatmapRow, error) {
	return r.q.GetActivityHeatmap(ctx, daysCount)
}
