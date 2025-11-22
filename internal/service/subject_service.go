package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type SubjectService interface {
	CreateSubject(ctx context.Context, name, colorHex string) (database.Subject, error)
	ListSubjects(ctx context.Context) ([]database.Subject, error)
	GetSubject(ctx context.Context, id string) (database.Subject, error)
	UpdateSubject(ctx context.Context, id, name, colorHex string) error
	DeleteSubject(ctx context.Context, id string) error
}

type SubjectManager struct {
	repo repository.SubjectRepository
}

func NewSubjectManager(repo repository.SubjectRepository) *SubjectManager {
	return &SubjectManager{repo: repo}
}

func (s *SubjectManager) CreateSubject(ctx context.Context, name, colorHex string) (database.Subject, error) {
	id := uuid.New().String()

	var color sql.NullString
	if colorHex != "" {
		color = sql.NullString{String: colorHex, Valid: true}
	}

	return s.repo.CreateSubject(ctx, database.CreateSubjectParams{
		ID:       id,
		Name:     name,
		ColorHex: color,
	})
}

func (s *SubjectManager) ListSubjects(ctx context.Context) ([]database.Subject, error) {
	return s.repo.ListSubjects(ctx)
}

func (s *SubjectManager) GetSubject(ctx context.Context, id string) (database.Subject, error) {
	return s.repo.GetSubject(ctx, id)
}

func (s *SubjectManager) UpdateSubject(ctx context.Context, id, name, colorHex string) error {
	var color sql.NullString
	if colorHex != "" {
		color = sql.NullString{String: colorHex, Valid: true}
	}

	return s.repo.UpdateSubject(ctx, database.UpdateSubjectParams{
		Name:     name,
		ColorHex: color,
		ID:       id,
	})
}

func (s *SubjectManager) DeleteSubject(ctx context.Context, id string) error {
	return s.repo.DeleteSubject(ctx, id)
}
