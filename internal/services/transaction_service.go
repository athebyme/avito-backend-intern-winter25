package services

//
//type TransactionService struct {
//	transactionRepo storage.TransactionRepository
//	userRepo        storage.UserRepository
//	db              *sql.DB
//}
//
//func (s *TransactionService) TransferCoins(fromUserID, toUserID int64, amount int) error {
//	if amount <= 0 {
//		return errors.New("amount must be positive")
//	}
//
//	tx, err := s.db.Begin()
//	if err != nil {
//		return err
//	}
//	defer tx.Rollback()
//
//	fromUser, err := s.userRepo.FindByIDForUpdate(tx, fromUserID)
//	if err != nil {
//		return err
//	}
//
//	if fromUser.Coins < amount {
//		return errors.New("insufficient coins")
//	}
//
//	toUser, err := s.userRepo.FindByIDForUpdate(tx, toUserID)
//	if err != nil {
//		return errors.New("recipient not found")
//	}
//
//	fromUser.Coins -= amount
//	toUser.Coins += amount
//
//	if err := s.userRepo.Update(tx, fromUser); err != nil {
//		return err
//	}
//	if err := s.userRepo.Update(tx, toUser); err != nil {
//		return err
//	}
//
//	transaction := &models.CoinTransaction{
//		FromUserID: fromUserID,
//		ToUserID:   toUserID,
//		Amount:     amount,
//		CreatedAt:  time.Now(),
//	}
//
//	if err := s.transactionRepo.Create(tx, transaction); err != nil {
//		return err
//	}
//
//	return tx.Commit()
//}
