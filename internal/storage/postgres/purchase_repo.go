package postgres

import (
	"avito-backend-intern-winter25/internal/models"
	"database/sql"
	"time"
)

type PurchaseRepository struct {
	db *sql.DB
}

func NewPurchaseRepository(db *sql.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

func (r *PurchaseRepository) Create(purchase *models.Purchase) error {
	query := `
        INSERT INTO purchases (user_id, item, price, purchase_date)
        VALUES ($1, $2, $3, $4) RETURNING id
    `
	if purchase.PurchaseDate.IsZero() {
		purchase.PurchaseDate = time.Now()
	}
	return r.db.QueryRow(query, purchase.UserID, purchase.Item, purchase.Price, purchase.PurchaseDate).Scan(&purchase.ID)
}

func (r *PurchaseRepository) GetByUser(userID int64) ([]*models.Purchase, error) {
	query := `
        SELECT id, user_id, item, price, purchase_date
        FROM purchases
        WHERE user_id = $1
        ORDER BY purchase_date DESC
    `
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var purchases []*models.Purchase
	for rows.Next() {
		var p models.Purchase
		if err := rows.Scan(&p.ID, &p.UserID, &p.Item, &p.Price, &p.PurchaseDate); err != nil {
			return nil, err
		}
		purchases = append(purchases, &p)
	}
	return purchases, nil
}
