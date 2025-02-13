package postgres

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"avito-backend-intern-winter25/internal/storage"
	"context"
	"database/sql"
	"errors"
	"time"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *sql.Tx, transaction *domain.CoinTransaction) error {
	if tx == nil {
		return errors.New("tx is nil")
	}
	query := `
        INSERT INTO coin_transactions (from_user_id, to_user_id, amount, created_at)
        VALUES ($1, $2, $3, $4) RETURNING id
    `
	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = time.Now()
	}
	return tx.QueryRowContext(ctx, query, transaction.FromUserID, transaction.ToUserID, transaction.Amount, transaction.CreatedAt).Scan(&transaction.ID)
}

func (r *TransactionRepository) GetSentTransactions(ctx context.Context, userID int64) ([]*domain.CoinTransaction, error) {
	if userID < 0 {
		return nil, storage.ErrInvalidUserId
	}

	query := `
        SELECT id, from_user_id, to_user_id, amount, created_at
        FROM coin_transactions
        WHERE from_user_id = $1
        ORDER BY created_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.CoinTransaction
	for rows.Next() {
		var t domain.CoinTransaction
		if err := rows.Scan(&t.ID, &t.FromUserID, &t.ToUserID, &t.Amount, &t.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, &t)
	}
	return transactions, nil
}

func (r *TransactionRepository) GetReceivedTransactions(ctx context.Context, userID int64) ([]*domain.CoinTransaction, error) {
	query := `
        SELECT id, from_user_id, to_user_id, amount, created_at
        FROM coin_transactions
        WHERE to_user_id = $1
        ORDER BY created_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.CoinTransaction
	for rows.Next() {
		var t domain.CoinTransaction
		if err := rows.Scan(&t.ID, &t.FromUserID, &t.ToUserID, &t.Amount, &t.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, &t)
	}
	return transactions, nil
}
