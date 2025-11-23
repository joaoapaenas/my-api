package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of service.UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, email, name, password string) (database.User, error) {
	args := m.Called(ctx, email, name, password)
	return args.Get(0).(database.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(database.User), args.Error(1)
}

func (m *MockUserService) UpdatePassword(ctx context.Context, email, oldPassword, newPassword string) error {
	args := m.Called(ctx, email, oldPassword, newPassword)
	return args.Error(0)
}

func TestUserHandler_CreateUser(t *testing.T) {
	mockSvc := new(MockUserService)
	h := handler.NewUserHandler(mockSvc)

	tests := []struct {
		name           string
		input          handler.CreateUserRequest
		mockReturnUser database.User
		mockReturnErr  error
		wantStatus     int
	}{
		{
			name: "Success",
			input: handler.CreateUserRequest{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			mockReturnUser: database.User{ID: "1", Email: "test@example.com"},
			mockReturnErr:  nil,
			wantStatus:     http.StatusCreated,
		},
		{
			name: "Validation Error",
			input: handler.CreateUserRequest{
				Email: "invalid-email",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			input: handler.CreateUserRequest{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			mockReturnErr: errors.New("db error"),
			wantStatus:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectation only if validation passes
			if tt.wantStatus != http.StatusBadRequest {
				mockSvc.On("CreateUser", mock.Anything, tt.input.Email, tt.input.Name, tt.input.Password).
					Return(tt.mockReturnUser, tt.mockReturnErr).
					Once()
			}

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			h.CreateUser(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantStatus == http.StatusCreated {
				var user database.User
				json.NewDecoder(rr.Body).Decode(&user)
				assert.Equal(t, tt.mockReturnUser.Email, user.Email)
			}
		})
	}
}
