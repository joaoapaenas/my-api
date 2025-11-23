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

// MockStudyCycleRepository is a mock implementation of repository.StudyCycleRepository
type MockStudyCycleRepository struct {
	mock.Mock
}

func (m *MockStudyCycleRepository) CreateStudyCycle(ctx context.Context, arg database.CreateStudyCycleParams) (database.StudyCycle, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.StudyCycle), args.Error(1)
}

func (m *MockStudyCycleRepository) GetActiveStudyCycle(ctx context.Context) (database.StudyCycle, error) {
	args := m.Called(ctx)
	return args.Get(0).(database.StudyCycle), args.Error(1)
}

func (m *MockStudyCycleRepository) GetStudyCycle(ctx context.Context, id string) (database.StudyCycle, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.StudyCycle), args.Error(1)
}

func (m *MockStudyCycleRepository) UpdateStudyCycle(ctx context.Context, arg database.UpdateStudyCycleParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockStudyCycleRepository) DeleteStudyCycle(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStudyCycleRepository) GetActiveCycleWithItems(ctx context.Context) ([]database.GetActiveCycleWithItemsRow, error) {
	args := m.Called(ctx)
	return args.Get(0).([]database.GetActiveCycleWithItemsRow), args.Error(1)
}

func TestStudyCycleManager_CreateStudyCycle(t *testing.T) {
	mockRepo := new(MockStudyCycleRepository)
	svc := service.NewStudyCycleManager(mockRepo)

	ctx := context.Background()
	name := "Cycle 1"
	description := "First Cycle"
	isActive := true

	mockRepo.On("CreateStudyCycle", ctx, mock.MatchedBy(func(arg database.CreateStudyCycleParams) bool {
		return arg.Name == name && arg.Description.String == description && arg.IsActive.Int64 == 1
	})).Return(database.StudyCycle{
		ID:          "cycle-uuid",
		Name:        name,
		Description: sql.NullString{String: description, Valid: true},
		IsActive:    sql.NullInt64{Int64: 1, Valid: true},
	}, nil)

	cycle, err := svc.CreateStudyCycle(ctx, name, description, isActive)

	assert.NoError(t, err)
	assert.Equal(t, name, cycle.Name)
	assert.Equal(t, int64(1), cycle.IsActive.Int64)
	mockRepo.AssertExpectations(t)
}

func TestStudyCycleManager_GetActiveStudyCycle(t *testing.T) {
	mockRepo := new(MockStudyCycleRepository)
	svc := service.NewStudyCycleManager(mockRepo)

	ctx := context.Background()
	expectedCycle := database.StudyCycle{ID: "active-uuid", Name: "Active Cycle", IsActive: sql.NullInt64{Int64: 1, Valid: true}}

	mockRepo.On("GetActiveStudyCycle", ctx).Return(expectedCycle, nil)

	cycle, err := svc.GetActiveStudyCycle(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Active Cycle", cycle.Name)
	mockRepo.AssertExpectations(t)
}
