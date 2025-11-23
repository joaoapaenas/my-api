package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailTaken   = errors.New("email already taken")
)

type UserService interface {
	CreateUser(ctx context.Context, email, name, password string) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
}

// UserManager implements UserService
type UserManager struct {
	repo repository.UserRepository
}

func NewUserManager(repo repository.UserRepository) *UserManager {
	return &UserManager{repo: repo}
}

func (s *UserManager) CreateUser(ctx context.Context, email, name, password string) (database.User, error) {
	// Logic: Generate UUID here, not in the handler
	id := uuid.New().String()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return database.User{}, err
	}

	user, err := s.repo.CreateUser(ctx, database.CreateUserParams{
		ID:           id,
		Email:        email,
		Name:         name,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		// In a real app, check for specific DB errors (like unique constraint violation)
		// and return ErrEmailTaken. For now, we return the raw error.
		return database.User{}, err
	}
	return user, nil
}

func (s *UserManager) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		// Assuming standard sql.ErrNoRows check happens here or in repo
		// Ideally, you map sql.ErrNoRows -> ErrUserNotFound here
		return database.User{}, err
	}
	return user, nil
}
