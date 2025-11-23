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

// MockStudySessionRepository is a mock implementation of repository.StudySessionRepository
type MockStudySessionRepository struct {
	mock.Mock
}

func (m *MockStudySessionRepository) CreateStudySession(ctx context.Context, arg database.CreateStudySessionParams) (database.StudySession, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.StudySession), args.Error(1)
}

func (m *MockStudySessionRepository) UpdateSessionDuration(ctx context.Context, arg database.UpdateSessionDurationParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockStudySessionRepository) GetStudySession(ctx context.Context, id string) (database.StudySession, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.StudySession), args.Error(1)
}

func (m *MockStudySessionRepository) DeleteStudySession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStudySessionRepository) GetOpenSession(ctx context.Context) (database.GetOpenSessionRow, error) {
	args := m.Called(ctx)
	return args.Get(0).(database.GetOpenSessionRow), args.Error(1)
}

func TestStudySessionManager_CreateStudySession(t *testing.T) {
	mockRepo := new(MockStudySessionRepository)
	svc := service.NewStudySessionManager(mockRepo)

	ctx := context.Background()
	subjectID := "subject-uuid"
	cycleItemID := "item-uuid"
	startedAt := "2023-10-27T10:00:00Z"

	mockRepo.On("CreateStudySession", ctx, mock.MatchedBy(func(arg database.CreateStudySessionParams) bool {
		return arg.SubjectID == subjectID && arg.CycleItemID.String == cycleItemID && arg.StartedAt == startedAt
	})).Return(database.StudySession{
		ID:          "session-uuid",
		SubjectID:   subjectID,
		CycleItemID: sql.NullString{String: cycleItemID, Valid: true},
		StartedAt:   startedAt,
	}, nil)

	session, err := svc.CreateStudySession(ctx, subjectID, cycleItemID, startedAt)

	assert.NoError(t, err)
	assert.Equal(t, subjectID, session.SubjectID)
	assert.Equal(t, cycleItemID, session.CycleItemID.String)
	mockRepo.AssertExpectations(t)
}

func TestStudySessionManager_UpdateSessionDuration(t *testing.T) {
	mockRepo := new(MockStudySessionRepository)
	svc := service.NewStudySessionManager(mockRepo)

	ctx := context.Background()
	sessionID := "session-uuid"
	finishedAt := "2023-10-27T11:00:00Z"
	gross := 3600
	net := 3000
	notes := "Good session"

	mockRepo.On("UpdateSessionDuration", ctx, mock.MatchedBy(func(arg database.UpdateSessionDurationParams) bool {
		return arg.ID == sessionID && arg.FinishedAt.String == finishedAt && arg.GrossDurationSeconds.Int64 == int64(gross) && arg.NetDurationSeconds.Int64 == int64(net)
	})).Return(nil)

	err := svc.UpdateSessionDuration(ctx, sessionID, finishedAt, gross, net, notes)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
