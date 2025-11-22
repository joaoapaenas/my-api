package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type StudySessionService interface {
	CreateStudySession(ctx context.Context, subjectID, cycleItemID, startedAt string) (database.StudySession, error)
	UpdateSessionDuration(ctx context.Context, id, finishedAt string, grossSeconds, netSeconds int, notes string) error
	GetStudySession(ctx context.Context, id string) (database.StudySession, error)
	DeleteStudySession(ctx context.Context, id string) error
	GetOpenSession(ctx context.Context) (database.GetOpenSessionRow, error)
}

type StudySessionManager struct {
	repo repository.StudySessionRepository
}

func NewStudySessionManager(repo repository.StudySessionRepository) *StudySessionManager {
	return &StudySessionManager{repo: repo}
}

func (s *StudySessionManager) CreateStudySession(ctx context.Context, subjectID, cycleItemID, startedAt string) (database.StudySession, error) {
	id := uuid.New().String()

	var cycleItem sql.NullString
	if cycleItemID != "" {
		cycleItem = sql.NullString{String: cycleItemID, Valid: true}
	}

	return s.repo.CreateStudySession(ctx, database.CreateStudySessionParams{
		ID:          id,
		SubjectID:   subjectID,
		CycleItemID: cycleItem,
		StartedAt:   startedAt,
	})
}

func (s *StudySessionManager) UpdateSessionDuration(ctx context.Context, id, finishedAt string, grossSeconds, netSeconds int, notes string) error {
	var finished sql.NullString
	if finishedAt != "" {
		finished = sql.NullString{String: finishedAt, Valid: true}
	}

	var gross sql.NullInt64
	if grossSeconds > 0 {
		gross = sql.NullInt64{Int64: int64(grossSeconds), Valid: true}
	}

	var net sql.NullInt64
	if netSeconds > 0 {
		net = sql.NullInt64{Int64: int64(netSeconds), Valid: true}
	}

	var sessionNotes sql.NullString
	if notes != "" {
		sessionNotes = sql.NullString{String: notes, Valid: true}
	}

	return s.repo.UpdateSessionDuration(ctx, database.UpdateSessionDurationParams{
		FinishedAt:           finished,
		GrossDurationSeconds: gross,
		NetDurationSeconds:   net,
		Notes:                sessionNotes,
		ID:                   id,
	})
}

func (s *StudySessionManager) GetStudySession(ctx context.Context, id string) (database.StudySession, error) {
	return s.repo.GetStudySession(ctx, id)
}

func (s *StudySessionManager) DeleteStudySession(ctx context.Context, id string) error {
	return s.repo.DeleteStudySession(ctx, id)
}

func (s *StudySessionManager) GetOpenSession(ctx context.Context) (database.GetOpenSessionRow, error) {
	return s.repo.GetOpenSession(ctx)
}
