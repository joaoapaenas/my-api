package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type TopicRepository interface {
	CreateTopic(ctx context.Context, arg database.CreateTopicParams) (database.Topic, error)
	ListTopicsBySubject(ctx context.Context, subjectID string) ([]database.Topic, error)
}

type SQLTopicRepository struct {
	q database.Querier
}

func NewSQLTopicRepository(q database.Querier) *SQLTopicRepository {
	return &SQLTopicRepository{q: q}
}

func (r *SQLTopicRepository) CreateTopic(ctx context.Context, arg database.CreateTopicParams) (database.Topic, error) {
	return r.q.CreateTopic(ctx, arg)
}

func (r *SQLTopicRepository) ListTopicsBySubject(ctx context.Context, subjectID string) ([]database.Topic, error) {
	return r.q.ListTopicsBySubject(ctx, subjectID)
}
