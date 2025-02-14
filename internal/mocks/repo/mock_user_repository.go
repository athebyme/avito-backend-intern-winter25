package mocks_repo

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"context"
	"database/sql"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(*sql.Tx), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, tx *sql.Tx, user *domain.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByIDForUpdate(ctx context.Context, tx *sql.Tx, id int64) (*domain.User, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, tx *sql.Tx, user *domain.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}
