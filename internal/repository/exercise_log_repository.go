package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type ExerciseLogRepository interface {
	CreateExerciseLog(ctx context.Context, arg database.CreateExerciseLogParams) (database.ExerciseLog, error)
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
