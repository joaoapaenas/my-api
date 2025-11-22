package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type SessionPauseService interface {
	CreateSessionPause(ctx context.Context, sessionID, startedAt string) (database.SessionPause, error)
	EndSessionPause(ctx context.Context, id, endedAt string) error
}

type SessionPauseManager struct {
	repo repository.SessionPauseRepository
}

func NewSessionPauseManager(repo repository.SessionPauseRepository) *SessionPauseManager {
	return &SessionPauseManager{repo: repo}
}

func (s *SessionPauseManager) CreateSessionPause(ctx context.Context, sessionID, startedAt string) (database.SessionPause, error) {
	id := uuid.New().String()
	return s.repo.CreateSessionPause(ctx, database.CreateSessionPauseParams{
		ID:        id,
		SessionID: sessionID,
		StartedAt: startedAt,
	})
}

func (s *SessionPauseManager) EndSessionPause(ctx context.Context, id, endedAt string) error {
	var ended sql.NullString
	if endedAt != "" {
		ended = sql.NullString{String: endedAt, Valid: true}
	}

	return s.repo.EndSessionPause(ctx, database.EndSessionPauseParams{
		EndedAt: ended,
		ID:      id,
	})
}
