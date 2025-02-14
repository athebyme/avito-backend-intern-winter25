package storage

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"context"
	"database/sql"
	"errors"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository interface {
	Create(ctx context.Context, tx *sql.Tx, user *domain.User) error
	Update(ctx context.Context, tx *sql.Tx, user *domain.User) error
	FindByIDForUpdate(ctx context.Context, tx *sql.Tx, id int64) (*domain.User, error)

	FindByID(ctx context.Context, id int64) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)

	// BeginTx : вынес, дабы не завязывать сервис на sql.db и вынести транзакцию на уровень сервиса + CRUD !
	BeginTx(ctx context.Context) (*sql.Tx, error)
}
