package storage

import (
	"avito-backend-intern-winter25/internal/models"
)

type PurchaseRepository interface {
	Create(purchase *models.Purchase) error
	GetByUser(userID int64) ([]*models.Purchase, error)
}
