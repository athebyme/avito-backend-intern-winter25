package storage

import (
	"avito-backend-intern-winter25/internal/models"
	"errors"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id int64) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	Update(user *models.User) error
}
