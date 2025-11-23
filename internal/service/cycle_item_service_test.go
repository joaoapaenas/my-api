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

// MockCycleItemRepository is a mock implementation of repository.CycleItemRepository
type MockCycleItemRepository struct {
	mock.Mock
}

func (m *MockCycleItemRepository) CreateCycleItem(ctx context.Context, arg database.CreateCycleItemParams) (database.CycleItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.CycleItem), args.Error(1)
}

func (m *MockCycleItemRepository) ListCycleItems(ctx context.Context, cycleID string) ([]database.CycleItem, error) {
	args := m.Called(ctx, cycleID)
	return args.Get(0).([]database.CycleItem), args.Error(1)
}

func (m *MockCycleItemRepository) GetCycleItem(ctx context.Context, id string) (database.CycleItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.CycleItem), args.Error(1)
}

func (m *MockCycleItemRepository) UpdateCycleItem(ctx context.Context, arg database.UpdateCycleItemParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockCycleItemRepository) DeleteCycleItem(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCycleItemManager_CreateCycleItem(t *testing.T) {
	mockRepo := new(MockCycleItemRepository)
	svc := service.NewCycleItemManager(mockRepo)

	ctx := context.Background()
	cycleID := "cycle-uuid"
	subjectID := "subject-uuid"
	orderIndex := 1
	plannedDuration := 60

	mockRepo.On("CreateCycleItem", ctx, mock.MatchedBy(func(arg database.CreateCycleItemParams) bool {
		return arg.CycleID == cycleID && arg.SubjectID == subjectID && arg.OrderIndex == int64(orderIndex) && arg.PlannedDurationMinutes.Int64 == int64(plannedDuration)
	})).Return(database.CycleItem{
		ID:                     "item-uuid",
		CycleID:                cycleID,
		SubjectID:              subjectID,
		OrderIndex:             int64(orderIndex),
		PlannedDurationMinutes: sql.NullInt64{Int64: int64(plannedDuration), Valid: true},
	}, nil)

	item, err := svc.CreateCycleItem(ctx, cycleID, subjectID, orderIndex, plannedDuration)

	assert.NoError(t, err)
	assert.Equal(t, int64(orderIndex), item.OrderIndex)
	assert.Equal(t, int64(plannedDuration), item.PlannedDurationMinutes.Int64)
	mockRepo.AssertExpectations(t)
}

func TestCycleItemManager_ListCycleItems(t *testing.T) {
	mockRepo := new(MockCycleItemRepository)
	svc := service.NewCycleItemManager(mockRepo)

	ctx := context.Background()
	cycleID := "cycle-uuid"
	expectedItems := []database.CycleItem{
		{ID: "1", CycleID: cycleID, OrderIndex: 1},
		{ID: "2", CycleID: cycleID, OrderIndex: 2},
	}

	mockRepo.On("ListCycleItems", ctx, cycleID).Return(expectedItems, nil)

	items, err := svc.ListCycleItems(ctx, cycleID)

	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, int64(1), items[0].OrderIndex)
	mockRepo.AssertExpectations(t)
}
