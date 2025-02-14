package services

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"avito-backend-intern-winter25/internal/storage"
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrLackOfFundsOnAccount = errors.New("lack of funds on account")
)

type TransactionService struct {
	transactionRepo storage.TransactionRepository
	userRepo        storage.UserRepository
	db              *sql.DB
}

func NewTransactionService(
	db *sql.DB,
	userRepo storage.UserRepository,
	transactionRepo storage.TransactionRepository) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
		db:              db,
	}
}

func (s *TransactionService) TransferCoins(ctx context.Context, fromUserID, toUserID int64, amount int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	fromUser, err := s.userRepo.FindByIDForUpdate(ctx, tx, fromUserID)
	if err != nil {
		return err
	}
	toUser, err := s.userRepo.FindByIDForUpdate(ctx, tx, toUserID)
	if err != nil {
		return err
	}

	if fromUser.Coins < amount {
		return ErrLackOfFundsOnAccount
	}

	fromUser.Coins -= amount
	toUser.Coins += amount

	if err := s.userRepo.Update(ctx, tx, fromUser); err != nil {
		return err
	}
	if err := s.userRepo.Update(ctx, tx, toUser); err != nil {
		return err
	}

	transaction := &domain.CoinTransaction{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Amount:     amount,
		CreatedAt:  time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, tx, transaction); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *TransactionService) GetSentTransactions(ctx context.Context, userID int64) ([]*domain.CoinTransaction, error) {
	return s.transactionRepo.GetSentTransactions(ctx, userID)
}

func (s *TransactionService) GetReceivedTransactions(ctx context.Context, userID int64) ([]*domain.CoinTransaction, error) {
	return s.transactionRepo.GetReceivedTransactions(ctx, userID)
}
