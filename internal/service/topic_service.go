package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type TopicService interface {
	CreateTopic(ctx context.Context, subjectID, name string) (database.Topic, error)
	ListTopicsBySubject(ctx context.Context, subjectID string) ([]database.Topic, error)
}

type TopicManager struct {
	repo repository.TopicRepository
}

func NewTopicManager(repo repository.TopicRepository) *TopicManager {
	return &TopicManager{repo: repo}
}

func (s *TopicManager) CreateTopic(ctx context.Context, subjectID, name string) (database.Topic, error) {
	id := uuid.New().String()
	return s.repo.CreateTopic(ctx, database.CreateTopicParams{
		ID:        id,
		SubjectID: subjectID,
		Name:      name,
	})
}

func (s *TopicManager) ListTopicsBySubject(ctx context.Context, subjectID string) ([]database.Topic, error) {
	return s.repo.ListTopicsBySubject(ctx, subjectID)
}
