package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type CycleItemRepository interface {
	CreateCycleItem(ctx context.Context, arg database.CreateCycleItemParams) (database.CycleItem, error)
	ListCycleItems(ctx context.Context, cycleID string) ([]database.CycleItem, error)
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
