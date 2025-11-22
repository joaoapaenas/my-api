package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type SessionPauseRepository interface {
	CreateSessionPause(ctx context.Context, arg database.CreateSessionPauseParams) (database.SessionPause, error)
	EndSessionPause(ctx context.Context, arg database.EndSessionPauseParams) error
	GetSessionPause(ctx context.Context, id string) (database.SessionPause, error)
	DeleteSessionPause(ctx context.Context, id string) error
}

type SQLSessionPauseRepository struct {
	q database.Querier
}

func NewSQLSessionPauseRepository(q database.Querier) *SQLSessionPauseRepository {
	return &SQLSessionPauseRepository{q: q}
}

func (r *SQLSessionPauseRepository) CreateSessionPause(ctx context.Context, arg database.CreateSessionPauseParams) (database.SessionPause, error) {
	return r.q.CreateSessionPause(ctx, arg)
}

func (r *SQLSessionPauseRepository) EndSessionPause(ctx context.Context, arg database.EndSessionPauseParams) error {
	return r.q.EndSessionPause(ctx, arg)
}

func (r *SQLSessionPauseRepository) GetSessionPause(ctx context.Context, id string) (database.SessionPause, error) {
	return r.q.GetSessionPause(ctx, id)
}

func (r *SQLSessionPauseRepository) DeleteSessionPause(ctx context.Context, id string) error {
	return r.q.DeleteSessionPause(ctx, id)
}
