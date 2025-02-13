package services

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"avito-backend-intern-winter25/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

var (
	ErrInsufficientCoins = errors.New("insufficient coins")
)

type MerchService struct {
	merchRepo    storage.MerchRepository
	purchaseRepo storage.PurchaseRepository
	userRepo     storage.UserRepository
	db           *sql.DB
}

func NewMerchService(
	merchRepo storage.MerchRepository,
	purchaseRepo storage.PurchaseRepository,
	userRepo storage.UserRepository,
	db *sql.DB,
) *MerchService {
	return &MerchService{
		merchRepo:    merchRepo,
		purchaseRepo: purchaseRepo,
		userRepo:     userRepo,
		db:           db,
	}
}

func (s *MerchService) PurchaseItem(ctx context.Context, userID int64, itemName string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("rollback error: %v", err)
		}
	}()

	item, err := s.merchRepo.FindByName(ctx, itemName)
	if err != nil {
		return fmt.Errorf("merch not found: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.Coins < item.Price {
		return ErrInsufficientCoins
	}

	user.Coins -= item.Price
	if err := s.userRepo.Update(ctx, tx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if err := s.purchaseRepo.Create(ctx, tx, &domain.Purchase{
		UserID:       userID,
		Item:         item.Name,
		Price:        item.Price,
		PurchaseDate: time.Now(),
	}); err != nil {
		return fmt.Errorf("failed to create purchase: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}
