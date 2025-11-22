package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type ExerciseLogRepository interface {
	CreateExerciseLog(ctx context.Context, arg database.CreateExerciseLogParams) (database.ExerciseLog, error)
	GetExerciseLog(ctx context.Context, id string) (database.ExerciseLog, error)
	DeleteExerciseLog(ctx context.Context, id string) error
}

type SQLExerciseLogRepository struct {
	q database.Querier
}

func NewSQLExerciseLogRepository(q database.Querier) *SQLExerciseLogRepository {
	return &SQLExerciseLogRepository{q: q}
}

func (r *SQLExerciseLogRepository) CreateExerciseLog(ctx context.Context, arg database.CreateExerciseLogParams) (database.ExerciseLog, error) {
	return r.q.CreateExerciseLog(ctx, arg)
}

func (r *SQLExerciseLogRepository) GetExerciseLog(ctx context.Context, id string) (database.ExerciseLog, error) {
	return r.q.GetExerciseLog(ctx, id)
}

func (r *SQLExerciseLogRepository) DeleteExerciseLog(ctx context.Context, id string) error {
	return r.q.DeleteExerciseLog(ctx, id)
}
