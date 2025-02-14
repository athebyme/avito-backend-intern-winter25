package storage

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"context"
	"errors"
)

var (
	ErrMerchNameIsIncorrect = errors.New("MerchName is incorrect")
	ErrMerchNotFound        = errors.New("merch not found")
)

type MerchRepository interface {
	GetAllAvailableMerch(ctx context.Context) ([]*domain.Merch, error)
	FindByName(ctx context.Context, name string) (*domain.Merch, error)
}
