package services

import (
	"avito-backend-intern-winter25/internal/models"
	"avito-backend-intern-winter25/internal/storage"
	"database/sql"
	"errors"
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

func (s *MerchService) PurchaseItem(userID int64, itemName string) error {
	item, err := s.merchRepo.FindByName(itemName)
	if err != nil {
		return storage.ErrMerchNotFound
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			panic(err)
		}
	}(tx)

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if user.Coins < item.Price {
		return ErrInsufficientCoins
	}

	user.Coins -= item.Price
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	purchase := &models.Purchase{
		UserID:       userID,
		Item:         item.Name,
		Price:        item.Price,
		PurchaseDate: time.Now(),
	}

	if err := s.purchaseRepo.Create(purchase); err != nil {
		return err
	}

	return tx.Commit()
}
