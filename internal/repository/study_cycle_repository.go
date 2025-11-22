package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type StudyCycleRepository interface {
	CreateStudyCycle(ctx context.Context, arg database.CreateStudyCycleParams) (database.StudyCycle, error)
	GetActiveStudyCycle(ctx context.Context) (database.StudyCycle, error)
}

type SQLStudyCycleRepository struct {
	q database.Querier
}

func NewSQLStudyCycleRepository(q database.Querier) *SQLStudyCycleRepository {
	return &SQLStudyCycleRepository{q: q}
}

func (r *SQLStudyCycleRepository) CreateStudyCycle(ctx context.Context, arg database.CreateStudyCycleParams) (database.StudyCycle, error) {
	return r.q.CreateStudyCycle(ctx, arg)
}

func (r *SQLStudyCycleRepository) GetActiveStudyCycle(ctx context.Context) (database.StudyCycle, error) {
	return r.q.GetActiveStudyCycle(ctx)
}
