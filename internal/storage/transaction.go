package storage

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"context"
	"errors"
)

var (
	ErrInvalidUserID = errors.New("invalid user id")
)

type TransactionRepository interface {
	Create(ctx context.Context, tx Tx, transaction *domain.CoinTransaction) error
	GetSentTransactions(ctx context.Context, userID int64) ([]*domain.CoinTransaction, error)
	GetReceivedTransactions(ctx context.Context, userID int64) ([]*domain.CoinTransaction, error)
}
