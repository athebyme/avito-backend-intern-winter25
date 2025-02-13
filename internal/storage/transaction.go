package storage

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"context"
	"database/sql"
	"errors"
)

var (
	ErrInvalidUserId = errors.New("invalid user id")
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *sql.Tx, transaction *domain.CoinTransaction) error
	GetSentTransactions(ctx context.Context, tx *sql.Tx, userID int64) ([]*domain.CoinTransaction, error)
	GetReceivedTransactions(ctx context.Context, tx *sql.Tx, userID int64) ([]*domain.CoinTransaction, error)
}
