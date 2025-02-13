package storage

import (
	"avito-backend-intern-winter25/internal/models"
	"errors"
)

var (
	ErrMerchNameIsIncorrect = errors.New("MerchName is incorrect")
	ErrMerchNotFound        = errors.New("merch not found")
)

type MerchRepository interface {
	FindByName(name string) (models.Merch, error)
}
