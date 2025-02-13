package storage

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"context"
	"database/sql"
)

type PurchaseRepository interface {
	Create(ctx context.Context, tx *sql.Tx, purchase *domain.Purchase) error
	GetByUser(ctx context.Context, tx *sql.Tx, userID int64) ([]*domain.Purchase, error)
}
