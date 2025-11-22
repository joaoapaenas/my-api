package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type CycleItemRepository interface {
	CreateCycleItem(ctx context.Context, arg database.CreateCycleItemParams) (database.CycleItem, error)
	ListCycleItems(ctx context.Context, cycleID string) ([]database.CycleItem, error)
	GetCycleItem(ctx context.Context, id string) (database.CycleItem, error)
	UpdateCycleItem(ctx context.Context, arg database.UpdateCycleItemParams) error
	DeleteCycleItem(ctx context.Context, id string) error
}

type SQLCycleItemRepository struct {
	q database.Querier
}

func NewSQLCycleItemRepository(q database.Querier) *SQLCycleItemRepository {
	return &SQLCycleItemRepository{q: q}
}

func (r *SQLCycleItemRepository) CreateCycleItem(ctx context.Context, arg database.CreateCycleItemParams) (database.CycleItem, error) {
	return r.q.CreateCycleItem(ctx, arg)
}

func (r *SQLCycleItemRepository) ListCycleItems(ctx context.Context, cycleID string) ([]database.CycleItem, error) {
	return r.q.ListCycleItems(ctx, cycleID)
}

func (r *SQLCycleItemRepository) GetCycleItem(ctx context.Context, id string) (database.CycleItem, error) {
	return r.q.GetCycleItem(ctx, id)
}

func (r *SQLCycleItemRepository) UpdateCycleItem(ctx context.Context, arg database.UpdateCycleItemParams) error {
	return r.q.UpdateCycleItem(ctx, arg)
}

func (r *SQLCycleItemRepository) DeleteCycleItem(ctx context.Context, id string) error {
	return r.q.DeleteCycleItem(ctx, id)
}
