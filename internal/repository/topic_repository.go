package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type TopicRepository interface {
	CreateTopic(ctx context.Context, arg database.CreateTopicParams) (database.Topic, error)
	ListTopicsBySubject(ctx context.Context, subjectID string) ([]database.Topic, error)
	GetTopic(ctx context.Context, id string) (database.Topic, error)
	UpdateTopic(ctx context.Context, arg database.UpdateTopicParams) error
	DeleteTopic(ctx context.Context, id string) error
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

func (r *SQLTopicRepository) GetTopic(ctx context.Context, id string) (database.Topic, error) {
	return r.q.GetTopic(ctx, id)
}

func (r *SQLTopicRepository) UpdateTopic(ctx context.Context, arg database.UpdateTopicParams) error {
	return r.q.UpdateTopic(ctx, arg)
}

func (r *SQLTopicRepository) DeleteTopic(ctx context.Context, id string) error {
	return r.q.DeleteTopic(ctx, id)
}
