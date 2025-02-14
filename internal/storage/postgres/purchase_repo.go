package postgres

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"avito-backend-intern-winter25/pkg/errs"
	"context"
	"database/sql"
	"time"
)

type PurchaseRepository struct {
	db *sql.DB
}

func NewPurchaseRepository(db *sql.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

func (r *PurchaseRepository) Create(ctx context.Context, tx *sql.Tx, purchase *domain.Purchase) error {
	if tx == nil {
		return errs.ErrTransactionNotFound
	}

	query := `
        INSERT INTO purchases (user_id, item, price, purchase_date)
        VALUES ($1, $2, $3, $4) RETURNING id
    `
	if purchase.PurchaseDate.IsZero() {
		purchase.PurchaseDate = time.Now()
	}
	return tx.QueryRowContext(ctx, query, purchase.UserID, purchase.Item, purchase.Price, purchase.PurchaseDate).Scan(&purchase.ID)
}

func (r *PurchaseRepository) GetByUser(ctx context.Context, tx *sql.Tx, userID int64) ([]*domain.Purchase, error) {
	if tx == nil {
		return nil, errs.ErrTransactionNotFound
	}
	query := `
        SELECT id, user_id, item, price, purchase_date
        FROM purchases
        WHERE user_id = $1
        ORDER BY purchase_date DESC
    `
	rows, err := tx.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var purchases []*domain.Purchase
	for rows.Next() {
		var p domain.Purchase
		if err := rows.Scan(&p.ID, &p.UserID, &p.Item, &p.Price, &p.PurchaseDate); err != nil {
			return nil, err
		}
		purchases = append(purchases, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return purchases, nil
}
