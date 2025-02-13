package postgres

import (
	"avito-backend-intern-winter25/internal/models"
	"avito-backend-intern-winter25/internal/storage"
	"database/sql"
	"time"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(transaction *models.CoinTransaction) error {
	query := `
        INSERT INTO coin_transactions (from_user_id, to_user_id, amount, created_at)
        VALUES ($1, $2, $3, $4) RETURNING id
    `
	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = time.Now()
	}
	return r.db.QueryRow(query, transaction.FromUserID, transaction.ToUserID, transaction.Amount, transaction.CreatedAt).Scan(&transaction.ID)
}

func (r *TransactionRepository) GetSentTransactions(userID int64) ([]*models.CoinTransaction, error) {
	if userID < 0 {
		return nil, storage.ErrInvalidUserId
	}

	query := `
        SELECT id, from_user_id, to_user_id, amount, created_at
        FROM coin_transactions
        WHERE from_user_id = $1
        ORDER BY created_at DESC
    `
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*models.CoinTransaction
	for rows.Next() {
		var t models.CoinTransaction
		if err := rows.Scan(&t.ID, &t.FromUserID, &t.ToUserID, &t.Amount, &t.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, &t)
	}
	return transactions, nil
}

func (r *TransactionRepository) GetReceivedTransactions(userID int64) ([]*models.CoinTransaction, error) {
	query := `
        SELECT id, from_user_id, to_user_id, amount, created_at
        FROM coin_transactions
        WHERE to_user_id = $1
        ORDER BY created_at DESC
    `
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*models.CoinTransaction
	for rows.Next() {
		var t models.CoinTransaction
		if err := rows.Scan(&t.ID, &t.FromUserID, &t.ToUserID, &t.Amount, &t.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, &t)
	}
	return transactions, nil
}
