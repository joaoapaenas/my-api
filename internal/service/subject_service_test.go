package service_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSubjectRepository is a mock implementation of repository.SubjectRepository
type MockSubjectRepository struct {
	mock.Mock
}

func (m *MockSubjectRepository) CreateSubject(ctx context.Context, arg database.CreateSubjectParams) (database.Subject, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.Subject), args.Error(1)
}

func (m *MockSubjectRepository) ListSubjects(ctx context.Context, userID string) ([]database.Subject, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]database.Subject), args.Error(1)
}

func (m *MockSubjectRepository) GetSubject(ctx context.Context, id, userID string) (database.Subject, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(database.Subject), args.Error(1)
}

func (m *MockSubjectRepository) UpdateSubject(ctx context.Context, arg database.UpdateSubjectParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockSubjectRepository) DeleteSubject(ctx context.Context, id, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func TestSubjectManager_CreateSubject(t *testing.T) {
	mockRepo := new(MockSubjectRepository)
	svc := service.NewSubjectManager(mockRepo)

	ctx := context.Background()
	userID := "user-123"
	name := "Mathematics"
	colorHex := "#FF0000"

	// Expect CreateSubject to be called with UserID
	mockRepo.On("CreateSubject", ctx, mock.MatchedBy(func(arg database.CreateSubjectParams) bool {
		return arg.Name == name &&
			   arg.ColorHex.String == colorHex &&
			   arg.ColorHex.Valid &&
			   arg.UserID == userID
	})).Return(database.Subject{
		ID:       "uuid",
		UserID:   userID,
		Name:     name,
		ColorHex: sql.NullString{String: colorHex, Valid: true},
	}, nil)

	subject, err := svc.CreateSubject(ctx, userID, name, colorHex)

	assert.NoError(t, err)
	assert.Equal(t, name, subject.Name)
	assert.Equal(t, userID, subject.UserID)
	assert.Equal(t, colorHex, subject.ColorHex.String)
	mockRepo.AssertExpectations(t)
}

func TestSubjectManager_ListSubjects(t *testing.T) {
	mockRepo := new(MockSubjectRepository)
	svc := service.NewSubjectManager(mockRepo)

	ctx := context.Background()
	userID := "user-123"
	expected
	package service_test

import (
		"context"
		"database/sql"
		"testing"

		"github.com/joaoapaenas/my-api/internal/database"
		"github.com/joaoapaenas/my-api/internal/service"
		"github.com/stretchr/testify/assert"
		"github.com/stretchr/testify/mock"
)

// MockSubjectRepository is a mock implementation of repository.SubjectRepository
type MockSubjectRepository struct {
		mock.Mock
}

func (m *MockSubjectRepository) CreateSubject(ctx context.Context, arg database.CreateSubjectParams) (database.Subject, error) {
		args := m.Called(ctx, arg)
		return args.Get(0).(database.Subject), args.Error(1)
}

func (m *MockSubjectRepository) ListSubjects(ctx context.Context, userID string) ([]database.Subject, error) {
		args := m.Called(ctx, userID)
		return args.Get(0).([]database.Subject), args.Error(1)
}

func (m *MockSubjectRepository) GetSubject(ctx context.Context, id, userID string) (database.Subject, error) {
		args := m.Called(ctx, id, userID)
		return args.Get(0).(database.Subject), args.Error(1)
}

func (m *MockSubjectRepository) UpdateSubject(ctx context.Context, arg database.UpdateSubjectParams) error {
		args := m.Called(ctx, arg)
		return args.Error(0)
}

func (m *MockSubjectRepository) DeleteSubject(ctx context.Context, id, userID string) error {
		args := m.Called(ctx, id, userID)
		return args.Error(0)
}

func TestSubjectManager_CreateSubject(t *testing.T) {
		mockRepo := new(MockSubjectRepository)
		svc := service.NewSubjectManager(mockRepo)

		ctx := context.Background()
		userID := "user-123"
		name := "Mathematics"
		colorHex := "#FF0000"

		// Expect CreateSubject to be called with UserID
		mockRepo.On("CreateSubject", ctx, mock.MatchedBy(func(arg database.CreateSubjectParams) bool {
			return arg.Name == name &&
				   arg.ColorHex.String == colorHex &&
				   arg.ColorHex.Valid &&
				   arg.UserID == userID
		})).Return(database.Subject{
			ID:       "uuid",
			UserID:   userID,
			Name:     name,
			ColorHex: sql.NullString{String: colorHex, Valid: true},
		}, nil)

		subject, err := svc.CreateSubject(ctx, userID, name, colorHex)

		assert.NoError(t, err)
		assert.Equal(t, name, subject.Name)
		assert.Equal(t, userID, subject.UserID)
		assert.Equal(t, colorHex, subject.ColorHex.String)
		mockRepo.AssertExpectations(t)
}

func TestSubjectManager_ListSubjects(t *testing.T) {
		mockRepo := new(MockSubjectRepository)
		svc := service.NewSubjectManager(mockRepo)

		ctx := context.Background()
		userID := "user-123"
		expectedSubjects := []database.Subject{
			{ID: "1", UserID: userID, Name: "Math"},
			{ID: "2", UserID: userID, Name: "Physics"},
		}

		// Expect ListSubjects to be called with UserID
		mockRepo.On("ListSubjects", ctx, userID).Return(expectedSubjects, nil)

		subjects, err := svc.ListSubjects(ctx, userID)

		assert.NoError(t, err)
		assert.Len(t, subjects, 2)
		assert.Equal(t, "Math", subjects[0].Name)
		mockRepo.AssertExpectations(t)
}
