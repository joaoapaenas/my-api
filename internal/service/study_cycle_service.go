package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type StudyCycleService interface {
	CreateStudyCycle(ctx context.Context, name, description string, isActive bool) (database.StudyCycle, error)
	GetActiveStudyCycle(ctx context.Context) (database.StudyCycle, error)
}

type StudyCycleManager struct {
	repo repository.StudyCycleRepository
}

func NewStudyCycleManager(repo repository.StudyCycleRepository) *StudyCycleManager {
	return &StudyCycleManager{repo: repo}
}

func (s *StudyCycleManager) CreateStudyCycle(ctx context.Context, name, description string, isActive bool) (database.StudyCycle, error) {
	id := uuid.New().String()

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	var active sql.NullInt64
	if isActive {
		active = sql.NullInt64{Int64: 1, Valid: true}
	} else {
		active = sql.NullInt64{Int64: 0, Valid: true}
	}

	return s.repo.CreateStudyCycle(ctx, database.CreateStudyCycleParams{
		ID:          id,
		Name:        name,
		Description: desc,
		IsActive:    active,
	})
}

func (s *StudyCycleManager) GetActiveStudyCycle(ctx context.Context) (database.StudyCycle, error) {
	return s.repo.GetActiveStudyCycle(ctx)
}
