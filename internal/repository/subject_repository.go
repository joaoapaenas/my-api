package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type SubjectRepository interface {
	CreateSubject(ctx context.Context, arg database.CreateSubjectParams) (database.Subject, error)
	ListSubjects(ctx context.Context) ([]database.Subject, error)
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

func (r *SQLSubjectRepository) ListSubjects(ctx context.Context) ([]database.Subject, error) {
	return r.q.ListSubjects(ctx)
}
