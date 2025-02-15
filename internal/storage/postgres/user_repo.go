package postgres

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"avito-backend-intern-winter25/internal/storage"
	"avito-backend-intern-winter25/pkg/errs"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) BeginTx(ctx context.Context) (storage.Tx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (r *UserRepository) Create(ctx context.Context, tx storage.Tx, user *domain.User) error {
	if tx == nil {
		return errs.ErrTransactionNotFound
	}

	query := `
        INSERT INTO users (username, password_hash, coins, created_at)
        VALUES ($1, $2, $3, $4) RETURNING id
    `
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	return tx.QueryRowContext(ctx, query, user.Username, user.PasswordHash, user.Coins, user.CreatedAt).Scan(&user.ID)
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
        SELECT id, username, password_hash, coins, created_at
        FROM users WHERE id = $1
    `
	var row *sql.Row
	row = r.db.QueryRowContext(ctx, query, id)

	var user domain.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Coins, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
        SELECT id, username, password_hash, coins, created_at
        FROM users WHERE username = $1
    `
	row := r.db.QueryRowContext(ctx, query, username)
	var user domain.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Coins, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, tx storage.Tx, user *domain.User) error {

	query := `
        UPDATE users SET username=$1, password_hash=$2, coins=$3
        WHERE id=$4
    `

	var res sql.Result
	var err error

	if tx != nil {
		res, err = tx.ExecContext(ctx, query, user.Username, user.PasswordHash, user.Coins, user.ID)
	} else {
		// если транзакцию не передали ?? мб стоит создавать
		res, err = r.db.ExecContext(ctx, query, user.Coins, user.ID)
	}

	if err != nil {
		return fmt.Errorf("update user failed: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected error: %w", err)
	}

	if rowsAffected == 0 {
		return storage.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) FindByIDForUpdate(ctx context.Context, tx storage.Tx, id int64) (*domain.User, error) {
	if tx == nil {
		return nil, errs.ErrTransactionNotFound
	}
	query := `
        SELECT id, username, password_hash, coins, created_at
        FROM users
        WHERE id = $1
        FOR UPDATE
    `
	row := tx.QueryRowContext(ctx, query, id)

	var user domain.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Coins, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
