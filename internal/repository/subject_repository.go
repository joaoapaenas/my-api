package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type SubjectRepository interface {
	CreateSubject(ctx context.Context, arg database.CreateSubjectParams) (database.Subject, error)
	ListSubjects(ctx context.Context) ([]database.Subject, error)
	GetSubject(ctx context.Context, id string) (database.Subject, error)
	UpdateSubject(ctx context.Context, arg database.UpdateSubjectParams) error
	DeleteSubject(ctx context.Context, id string) error
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

func (r *SQLSubjectRepository) GetSubject(ctx context.Context, id string) (database.Subject, error) {
	return r.q.GetSubject(ctx, id)
}

func (r *SQLSubjectRepository) UpdateSubject(ctx context.Context, arg database.UpdateSubjectParams) error {
	return r.q.UpdateSubject(ctx, arg)
}

func (r *SQLSubjectRepository) DeleteSubject(ctx context.Context, id string) error {
	return r.q.DeleteSubject(ctx, id)
}
