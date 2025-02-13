package services

import (
	"avito-backend-intern-winter25/internal/models"
	"avito-backend-intern-winter25/internal/storage"
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

func (s *TransactionService) TransferCoins(fromUserID, toUserID int64, amount int) error {
	if amount < 0 {
		return ErrInvalidAmount
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	fromUser, err := s.userRepo.FindByID(fromUserID)
	if err != nil {
		return err
	}

	if fromUser.Coins < amount {
		return ErrLackOfFundsOnAccount
	}

	toUser, err := s.userRepo.FindByID(toUserID)
	if err != nil {
		return storage.ErrUserNotFound
	}

	fromUser.Coins -= amount
	toUser.Coins += amount

	if err := s.userRepo.Update(fromUser); err != nil {
		return err
	}
	if err := s.userRepo.Update(toUser); err != nil {
		return err
	}

	transaction := &models.CoinTransaction{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Amount:     amount,
		CreatedAt:  time.Now(),
	}

	if err := s.transactionRepo.Create(transaction); err != nil {
		return err
	}

	return tx.Commit()
}
