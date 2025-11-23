package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type SubjectService interface {
	CreateSubject(ctx context.Context, userID, name, colorHex string) (database.Subject, error)
	ListSubjects(ctx context.Context, userID string) ([]database.Subject, error)
	GetSubject(ctx context.Context, id, userID string) (database.Subject, error)
	UpdateSubject(ctx context.Context, id, userID, name, colorHex string) error
	DeleteSubject(ctx context.Context, id, userID string) error
}

type SubjectManager struct {
	repo repository.SubjectRepository
}

func NewSubjectManager(repo repository.SubjectRepository) *SubjectManager {
	return &SubjectManager{repo: repo}
}

func (s *SubjectManager) CreateSubject(ctx context.Context, userID, name, colorHex string) (database.Subject, error) {
	id := uuid.New().String()

	var color sql.NullString
	if colorHex != "" {
		color = sql.NullString{String: colorHex, Valid: true}
	}

	// We now pass userID into the Params
	return s.repo.CreateSubject(ctx, database.CreateSubjectParams{
		ID:       id,
		UserID:   userID,
		Name:     name,
		ColorHex: color,
	})
}

func (s *SubjectManager) ListSubjects(ctx context.Context, userID string) ([]database.Subject, error) {
	subjects, err := s.repo.ListSubjects(ctx, userID)
	if err != nil {
		return nil, err
	}
	// Defensive fix for JSON null vs []
	if subjects == nil {
		return []database.Subject{}, nil
	}
	return subjects, nil
}

func (s *SubjectManager) GetSubject(ctx context.Context, id, userID string) (database.Subject, error) {
	return s.repo.GetSubject(ctx, id, userID)
}

func (s *SubjectManager) UpdateSubject(ctx context.Context, id, userID, name, colorHex string) error {
	var color sql.NullString
	if colorHex != "" {
		color = sql.NullString{String: colorHex, Valid: true}
	}

	// We now pass userID into the Params to ensure the WHERE clause checks ownership
	return s.repo.UpdateSubject(ctx, database.UpdateSubjectParams{
		Name:     name,
		ColorHex: color,
		ID:       id,
		UserID:   userID,
	})
}

func (s *SubjectManager) DeleteSubject(ctx context.Context, id, userID string) error {
	return s.repo.DeleteSubject(ctx, id, userID)
}
