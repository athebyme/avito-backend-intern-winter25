package storage

import (
	"avito-backend-intern-winter25/internal/models"
)

type TransactionRepository interface {
	Create(transaction *models.CoinTransaction) error
	GetSentTransactions(userID int64) ([]*models.CoinTransaction, error)
	GetReceivedTransactions(userID int64) ([]*models.CoinTransaction, error)
}
