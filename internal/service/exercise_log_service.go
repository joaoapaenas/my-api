package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type ExerciseLogService interface {
	CreateExerciseLog(ctx context.Context, sessionID, subjectID, topicID string, questionsCount, correctCount int) (database.ExerciseLog, error)
}

type ExerciseLogManager struct {
	repo repository.ExerciseLogRepository
}

func NewExerciseLogManager(repo repository.ExerciseLogRepository) *ExerciseLogManager {
	return &ExerciseLogManager{repo: repo}
}

func (s *ExerciseLogManager) CreateExerciseLog(ctx context.Context, sessionID, subjectID, topicID string, questionsCount, correctCount int) (database.ExerciseLog, error) {
	id := uuid.New().String()

	var session sql.NullString
	if sessionID != "" {
		session = sql.NullString{String: sessionID, Valid: true}
	}

	var topic sql.NullString
	if topicID != "" {
		topic = sql.NullString{String: topicID, Valid: true}
	}

	return s.repo.CreateExerciseLog(ctx, database.CreateExerciseLogParams{
		ID:             id,
		SessionID:      session,
		SubjectID:      subjectID,
		TopicID:        topic,
		QuestionsCount: int64(questionsCount),
		CorrectCount:   int64(correctCount),
	})
}
