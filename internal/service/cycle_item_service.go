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
	GetCycleItem(ctx context.Context, id string) (database.CycleItem, error)
	UpdateCycleItem(ctx context.Context, id, subjectID string, orderIndex int, plannedDurationMinutes int) error
	DeleteCycleItem(ctx context.Context, id string) error
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

func (s *CycleItemManager) GetCycleItem(ctx context.Context, id string) (database.CycleItem, error) {
	return s.repo.GetCycleItem(ctx, id)
}

func (s *CycleItemManager) UpdateCycleItem(ctx context.Context, id, subjectID string, orderIndex int, plannedDurationMinutes int) error {
	var duration sql.NullInt64
	if plannedDurationMinutes > 0 {
		duration = sql.NullInt64{Int64: int64(plannedDurationMinutes), Valid: true}
	}

	return s.repo.UpdateCycleItem(ctx, database.UpdateCycleItemParams{
		SubjectID:              subjectID,
		OrderIndex:             int64(orderIndex),
		PlannedDurationMinutes: duration,
		ID:                     id,
	})
}

func (s *CycleItemManager) DeleteCycleItem(ctx context.Context, id string) error {
	return s.repo.DeleteCycleItem(ctx, id)
}
