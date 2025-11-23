package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type SubjectRepository interface {
	// Create now expects the UserID inside arg
	CreateSubject(ctx context.Context, arg database.CreateSubjectParams) (database.Subject, error)

	// List is now filtered by userID
	ListSubjects(ctx context.Context, userID string) ([]database.Subject, error)

	// Get is now filtered by userID (prevent accessing others' IDs)
	GetSubject(ctx context.Context, id, userID string) (database.Subject, error)

	// Update expects UserID inside arg to ensure ownership before update
	UpdateSubject(ctx context.Context, arg database.UpdateSubjectParams) error

	// Delete requires userID to ensure ownership
	DeleteSubject(ctx context.Context, id, userID string) error
}

type SQLSubjectRepository struct {
	q database.Querier
}

func NewSQLSubjectRepository(q database.Querier) *SQLSubjectRepository {
	return &SQLSubjectRepository{q: q}
}

func (r *SQLSubjectRepository) CreateSubject(ctx context.Context, arg database.CreateSubjectParams) (database.Subject, error) {
	return r.q.CreateSubject(ctx, arg)
}

func (r *SQLSubjectRepository) ListSubjects(ctx context.Context, userID string) ([]database.Subject, error) {
	return r.q.ListSubjects(ctx, userID)
}

func (r *SQLSubjectRepository) GetSubject(ctx context.Context, id, userID string) (database.Subject, error) {
	// FIX: Wrap arguments in GetSubjectParams
	return r.q.GetSubject(ctx, database.GetSubjectParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *SQLSubjectRepository) UpdateSubject(ctx context.Context, arg database.UpdateSubjectParams) error {
	return r.q.UpdateSubject(ctx, arg)
}

func (r *SQLSubjectRepository) DeleteSubject(ctx context.Context, id, userID string) error {
	// FIX: Wrap arguments in DeleteSubjectParams
	return r.q.DeleteSubject(ctx, database.DeleteSubjectParams{
		ID:     id,
		UserID: userID,
	})
}
