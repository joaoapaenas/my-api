package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type UserRepository interface {
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
	UpdateUserPassword(ctx context.Context, id, passwordHash string) error
}

type SQLUserRepository struct {
	q database.Querier
}

func NewSQLUserRepository(q database.Querier) *SQLUserRepository {
	return &SQLUserRepository{q: q}
}

func (r *SQLUserRepository) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	return r.q.CreateUser(ctx, arg)
}

func (r *SQLUserRepository) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

func (r *SQLUserRepository) UpdateUserPassword(ctx context.Context, id, passwordHash string) error {
	return r.q.UpdateUserPassword(ctx, database.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: passwordHash,
	})
}
