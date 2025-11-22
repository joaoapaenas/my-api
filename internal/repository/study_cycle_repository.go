package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type StudyCycleRepository interface {
	CreateStudyCycle(ctx context.Context, arg database.CreateStudyCycleParams) (database.StudyCycle, error)
	GetActiveStudyCycle(ctx context.Context) (database.StudyCycle, error)
	GetStudyCycle(ctx context.Context, id string) (database.StudyCycle, error)
	UpdateStudyCycle(ctx context.Context, arg database.UpdateStudyCycleParams) error
	DeleteStudyCycle(ctx context.Context, id string) error
	GetActiveCycleWithItems(ctx context.Context) ([]database.GetActiveCycleWithItemsRow, error)
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

func (r *SQLStudyCycleRepository) GetStudyCycle(ctx context.Context, id string) (database.StudyCycle, error) {
	return r.q.GetStudyCycle(ctx, id)
}

func (r *SQLStudyCycleRepository) UpdateStudyCycle(ctx context.Context, arg database.UpdateStudyCycleParams) error {
	return r.q.UpdateStudyCycle(ctx, arg)
}

func (r *SQLStudyCycleRepository) DeleteStudyCycle(ctx context.Context, id string) error {
	return r.q.DeleteStudyCycle(ctx, id)
}

func (r *SQLStudyCycleRepository) GetActiveCycleWithItems(ctx context.Context) ([]database.GetActiveCycleWithItemsRow, error) {
	return r.q.GetActiveCycleWithItems(ctx)
}
