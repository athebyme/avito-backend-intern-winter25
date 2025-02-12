package postgres

import (
	"avito-backend-intern-winter25/internal/models"
	"avito-backend-intern-winter25/internal/storage"
	"database/sql"
	"errors"
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

func (r *UserRepository) Create(user *models.User) error {
	query := `
        INSERT INTO users (username, password_hash, coins, created_at)
        VALUES ($1, $2, $3, $4) RETURNING id
    `
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	return r.db.QueryRow(query, user.Username, user.PasswordHash, user.Coins, user.CreatedAt).Scan(&user.ID)
}

func (r *UserRepository) FindByID(id int64) (*models.User, error) {
	query := `
        SELECT id, username, password_hash, coins, created_at
        FROM users WHERE id = $1
    `
	row := r.db.QueryRow(query, id)
	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Coins, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	query := `
        SELECT id, username, password_hash, coins, created_at
        FROM users WHERE username = $1
    `
	row := r.db.QueryRow(query, username)
	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Coins, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	query := `
        UPDATE users SET username=$1, password_hash=$2, coins=$3
        WHERE id=$4
    `
	res, err := r.db.Exec(query, user.Username, user.PasswordHash, user.Coins, user.ID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
