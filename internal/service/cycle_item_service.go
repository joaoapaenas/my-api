package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type CycleItemService interface {
	CreateCycleItem(ctx context.Context, cycleID, subjectID string, orderIndex int, plannedDurationMinutes int) (database.CycleItem, error)
	ListCycleItems(ctx context.Context, cycleID string) ([]database.CycleItem, error)
}

type CycleItemManager struct {
	repo repository.CycleItemRepository
}

func NewCycleItemManager(repo repository.CycleItemRepository) *CycleItemManager {
	return &CycleItemManager{repo: repo}
}

func (s *CycleItemManager) CreateCycleItem(ctx context.Context, cycleID, subjectID string, orderIndex int, plannedDurationMinutes int) (database.CycleItem, error) {
	id := uuid.New().String()

	var duration sql.NullInt64
	if plannedDurationMinutes > 0 {
		duration = sql.NullInt64{Int64: int64(plannedDurationMinutes), Valid: true}
	}

	return s.repo.CreateCycleItem(ctx, database.CreateCycleItemParams{
		ID:                     id,
		CycleID:                cycleID,
		SubjectID:              subjectID,
		OrderIndex:             int64(orderIndex),
		PlannedDurationMinutes: duration,
	})
}

func (s *CycleItemManager) ListCycleItems(ctx context.Context, cycleID string) ([]database.CycleItem, error) {
	return s.repo.ListCycleItems(ctx, cycleID)
}
