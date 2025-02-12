package services

import (
	"avito-backend-intern-winter25/internal/storage"
	"database/sql"
)

type MerchService struct {
	// merchRepo    storage.MerchRepository
	purchaseRepo storage.PurchaseRepository
	userRepo     storage.UserRepository
	db           *sql.DB
}

//
//func (s *MerchService) PurchaseItem(userID int64, itemName string) error {
//	item, err := s.merchRepo.GetByName(itemName)
//	if err != nil {
//		return errors.New("item not found")
//	}
//
//	tx, err := s.db.Begin()
//	if err != nil {
//		return err
//	}
//	defer tx.Rollback()
//
//	user, err := s.userRepo.FindByIDForUpdate(tx, userID)
//	if err != nil {
//		return err
//	}
//
//	if user.Coins < item.Price {
//		return errors.New("insufficient coins")
//	}
//
//	user.Coins -= item.Price
//	if err := s.userRepo.Update(tx, user); err != nil {
//		return err
//	}
//
//	purchase := &models.Purchase{
//		UserID:       userID,
//		Item:         item.Name,
//		Price:        item.Price,
//		PurchaseDate: time.Now(),
//	}
//
//	if err := s.purchaseRepo.Create(tx, purchase); err != nil {
//		return err
//	}
//
//	return tx.Commit()
//}
