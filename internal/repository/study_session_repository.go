package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type StudySessionRepository interface {
	CreateStudySession(ctx context.Context, arg database.CreateStudySessionParams) (database.StudySession, error)
	UpdateSessionDuration(ctx context.Context, arg database.UpdateSessionDurationParams) error
	GetStudySession(ctx context.Context, id string) (database.StudySession, error)
	DeleteStudySession(ctx context.Context, id string) error
}

type SQLStudySessionRepository struct {
	q database.Querier
}

func NewSQLStudySessionRepository(q database.Querier) *SQLStudySessionRepository {
	return &SQLStudySessionRepository{q: q}
}

func (r *SQLStudySessionRepository) CreateStudySession(ctx context.Context, arg database.CreateStudySessionParams) (database.StudySession, error) {
	return r.q.CreateStudySession(ctx, arg)
}

func (r *SQLStudySessionRepository) UpdateSessionDuration(ctx context.Context, arg database.UpdateSessionDurationParams) error {
	return r.q.UpdateSessionDuration(ctx, arg)
}

func (r *SQLStudySessionRepository) GetStudySession(ctx context.Context, id string) (database.StudySession, error) {
	return r.q.GetStudySession(ctx, id)
}

func (r *SQLStudySessionRepository) DeleteStudySession(ctx context.Context, id string) error {
	return r.q.DeleteStudySession(ctx, id)
}
