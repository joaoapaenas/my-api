package service_test

import (
	"context"
	"testing"

	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTopicRepository is a mock implementation of repository.TopicRepository
type MockTopicRepository struct {
	mock.Mock
}

func (m *MockTopicRepository) CreateTopic(ctx context.Context, arg database.CreateTopicParams) (database.Topic, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.Topic), args.Error(1)
}

func (m *MockTopicRepository) ListTopicsBySubject(ctx context.Context, subjectID string) ([]database.Topic, error) {
	args := m.Called(ctx, subjectID)
	return args.Get(0).([]database.Topic), args.Error(1)
}

func (m *MockTopicRepository) GetTopic(ctx context.Context, id string) (database.Topic, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.Topic), args.Error(1)
}

func (m *MockTopicRepository) UpdateTopic(ctx context.Context, arg database.UpdateTopicParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockTopicRepository) DeleteTopic(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTopicManager_CreateTopic(t *testing.T) {
	mockRepo := new(MockTopicRepository)
	svc := service.NewTopicManager(mockRepo)

	ctx := context.Background()
	subjectID := "subject-uuid"
	name := "Algebra"

	mockRepo.On("CreateTopic", ctx, mock.MatchedBy(func(arg database.CreateTopicParams) bool {
		return arg.SubjectID == subjectID && arg.Name == name
	})).Return(database.Topic{
		ID:        "topic-uuid",
		SubjectID: subjectID,
		Name:      name,
	}, nil)

	topic, err := svc.CreateTopic(ctx, subjectID, name)

	assert.NoError(t, err)
	assert.Equal(t, name, topic.Name)
	assert.Equal(t, subjectID, topic.SubjectID)
	mockRepo.AssertExpectations(t)
}

func TestTopicManager_GetTopic(t *testing.T) {
	mockRepo := new(MockTopicRepository)
	svc := service.NewTopicManager(mockRepo)

	ctx := context.Background()
	topicID := "topic-uuid"
	expectedTopic := database.Topic{ID: topicID, Name: "Algebra"}

	mockRepo.On("GetTopic", ctx, topicID).Return(expectedTopic, nil)

	topic, err := svc.GetTopic(ctx, topicID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTopic.Name, topic.Name)
	mockRepo.AssertExpectations(t)
}
