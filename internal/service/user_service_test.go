package service_test

import (
	"context"
	"testing"

	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(database.User), args.Error(1)
}

func TestUserManager_CreateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := service.NewUserManager(mockRepo)

	ctx := context.Background()
	email := "test@example.com"
	name := "Test User"
	password := "password123"

	mockRepo.On("CreateUser", ctx, mock.MatchedBy(func(arg database.CreateUserParams) bool {
		// Verify arguments
		if arg.Email != email || arg.Name != name {
			return false
		}
		// Verify password is hashed
		err := bcrypt.CompareHashAndPassword([]byte(arg.PasswordHash), []byte(password))
		return err == nil
	})).Return(database.User{
		ID:    "uuid",
		Email: email,
		Name:  name,
	}, nil)

	user, err := svc.CreateUser(ctx, email, name, password)

	assert.NoError(t, err)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, name, user.Name)
	mockRepo.AssertExpectations(t)
}
