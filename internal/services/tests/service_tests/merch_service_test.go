package service_tests

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"avito-backend-intern-winter25/internal/services"
	"avito-backend-intern-winter25/internal/services/mocks"
	"avito-backend-intern-winter25/internal/storage"
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMerchService_PurchaseItem_Success(t *testing.T) {
	// arrange
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	merchRepo := new(mocks.MockMerchRepository)
	purchaseRepo := new(mocks.MockPurchaseRepository)
	userRepo := new(mocks.MockUserRepository)

	mockDB.ExpectBegin()
	mockDB.ExpectCommit()

	userID := int64(1)
	itemName := "T-Shirt"
	item := &domain.Merch{
		ID:    1,
		Name:  itemName,
		Price: 100,
	}
	user := &domain.User{
		ID:    userID,
		Coins: 200,
	}

	// act
	merchRepo.On("FindByName", mock.Anything, itemName).Return(item, nil)
	userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == userID && u.Coins == 100
	})).Return(nil)
	purchaseRepo.On("Create", mock.Anything, mock.Anything, mock.MatchedBy(func(p *domain.Purchase) bool {
		return p.UserID == userID && p.Item == itemName && p.Price == 100
	})).Return(nil)

	service := services.NewMerchService(merchRepo, purchaseRepo, userRepo, db)

	err = service.PurchaseItem(context.Background(), userID, itemName)

	// assert
	assert.NoError(t, err)
	merchRepo.AssertExpectations(t)
	purchaseRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestMerchService_PurchaseItem_InsufficientCoins(t *testing.T) {
	// arrange
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	merchRepo := new(mocks.MockMerchRepository)
	purchaseRepo := new(mocks.MockPurchaseRepository)
	userRepo := new(mocks.MockUserRepository)

	mockDB.ExpectBegin()
	mockDB.ExpectRollback()

	userID := int64(1)
	itemName := "Expensive T-Shirt"
	item := &domain.Merch{
		ID:    1,
		Name:  itemName,
		Price: 200,
	}
	user := &domain.User{
		ID:    userID,
		Coins: 100,
	}

	// act
	merchRepo.On("FindByName", mock.Anything, itemName).Return(item, nil)
	userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)

	service := services.NewMerchService(merchRepo, purchaseRepo, userRepo, db)

	err = service.PurchaseItem(context.Background(), userID, itemName)

	// assert
	assert.ErrorIs(t, err, services.ErrInsufficientCoins)
	merchRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestMerchService_PurchaseItem_ItemNotFound(t *testing.T) {
	// arrange
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	merchRepo := new(mocks.MockMerchRepository)
	purchaseRepo := new(mocks.MockPurchaseRepository)
	userRepo := new(mocks.MockUserRepository)

	mockDB.ExpectBegin()
	mockDB.ExpectRollback()

	userID := int64(1)
	itemName := "Non-existent Item"

	// ACT
	merchRepo.On("FindByName", mock.Anything, itemName).Return(nil, storage.ErrMerchNotFound)

	service := services.NewMerchService(merchRepo, purchaseRepo, userRepo, db)

	err = service.PurchaseItem(context.Background(), userID, itemName)

	// assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrMerchNotFound) ||
		errors.Is(errors.Unwrap(err), storage.ErrMerchNotFound))
	merchRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestMerchService_PurchaseItem_UserNotFound(t *testing.T) {
	// arrange
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	merchRepo := new(mocks.MockMerchRepository)
	purchaseRepo := new(mocks.MockPurchaseRepository)
	userRepo := new(mocks.MockUserRepository)

	mockDB.ExpectBegin()
	mockDB.ExpectRollback()

	userID := int64(999)
	itemName := "T-Shirt"
	item := &domain.Merch{
		ID:    1,
		Name:  itemName,
		Price: 100,
	}

	// act
	merchRepo.On("FindByName", mock.Anything, itemName).Return(item, nil)
	userRepo.On("FindByID", mock.Anything, userID).Return(nil, sql.ErrNoRows)

	service := services.NewMerchService(merchRepo, purchaseRepo, userRepo, db)

	err = service.PurchaseItem(context.Background(), userID, itemName)

	// assert
	assert.Error(t, err)
	merrs := errors.Unwrap(err)
	assert.True(t, errors.Is(merrs, sql.ErrNoRows))
	merchRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestMerchService_PurchaseItem_UpdateUserError(t *testing.T) {
	// arrange
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	merchRepo := new(mocks.MockMerchRepository)
	purchaseRepo := new(mocks.MockPurchaseRepository)
	userRepo := new(mocks.MockUserRepository)

	mockDB.ExpectBegin()
	mockDB.ExpectRollback()

	userID := int64(1)
	itemName := "T-Shirt"
	item := &domain.Merch{
		ID:    1,
		Name:  itemName,
		Price: 100,
	}
	user := &domain.User{
		ID:    userID,
		Coins: 200,
	}
	updateErr := errors.New("update failed")

	// act
	merchRepo.On("FindByName", mock.Anything, itemName).Return(item, nil)
	userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == userID && u.Coins == 100
	})).Return(updateErr)

	service := services.NewMerchService(merchRepo, purchaseRepo, userRepo, db)

	err = service.PurchaseItem(context.Background(), userID, itemName)

	// assert
	assert.Error(t, err)
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		assert.Equal(t, updateErr, unwrapped)
	}
	merchRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestMerchService_PurchaseItem_CreatePurchaseError(t *testing.T) {
	// arrange
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	merchRepo := new(mocks.MockMerchRepository)
	purchaseRepo := new(mocks.MockPurchaseRepository)
	userRepo := new(mocks.MockUserRepository)

	mockDB.ExpectBegin()
	mockDB.ExpectRollback()

	userID := int64(1)
	itemName := "T-Shirt"
	item := &domain.Merch{
		ID:    1,
		Name:  itemName,
		Price: 100,
	}
	user := &domain.User{
		ID:    userID,
		Coins: 200,
	}
	createErr := errors.New("create failed")

	// act
	merchRepo.On("FindByName", mock.Anything, itemName).Return(item, nil)
	userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == userID && u.Coins == 100
	})).Return(nil)
	purchaseRepo.On("Create", mock.Anything, mock.Anything, mock.MatchedBy(func(p *domain.Purchase) bool {
		return p.UserID == userID && p.Item == itemName && p.Price == 100
	})).Return(createErr)

	service := services.NewMerchService(merchRepo, purchaseRepo, userRepo, db)

	err = service.PurchaseItem(context.Background(), userID, itemName)

	// assert
	assert.Error(t, err)
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		assert.Equal(t, createErr, unwrapped)
	}
	merchRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	purchaseRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestMerchService_PurchaseItem_CommitError(t *testing.T) {
	// arrange
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	merchRepo := new(mocks.MockMerchRepository)
	purchaseRepo := new(mocks.MockPurchaseRepository)
	userRepo := new(mocks.MockUserRepository)

	mockDB.ExpectBegin()
	commitErr := errors.New("commit failed")
	mockDB.ExpectCommit().WillReturnError(commitErr)

	userID := int64(1)
	itemName := "T-Shirt"
	item := &domain.Merch{
		ID:    1,
		Name:  itemName,
		Price: 100,
	}
	user := &domain.User{
		ID:    userID,
		Coins: 200,
	}

	// act
	merchRepo.On("FindByName", mock.Anything, itemName).Return(item, nil)
	userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == userID && u.Coins == 100
	})).Return(nil)
	purchaseRepo.On("Create", mock.Anything, mock.Anything, mock.MatchedBy(func(p *domain.Purchase) bool {
		return p.UserID == userID && p.Item == itemName && p.Price == 100
	})).Return(nil)

	service := services.NewMerchService(merchRepo, purchaseRepo, userRepo, db)

	err = service.PurchaseItem(context.Background(), userID, itemName)

	// assert
	assert.Error(t, err)
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		assert.Equal(t, commitErr, unwrapped)
	}
	merchRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	purchaseRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestMerchService_GetPurchasesByUser_Success(t *testing.T) {
	// arrange
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	merchRepo := new(mocks.MockMerchRepository)
	purchaseRepo := new(mocks.MockPurchaseRepository)
	userRepo := new(mocks.MockUserRepository)

	mockDB.ExpectBegin()

	userID := int64(1)
	purchases := []*domain.Purchase{
		{ID: 1, UserID: userID, Item: "T-Shirt", Price: 100, PurchaseDate: time.Now()},
		{ID: 2, UserID: userID, Item: "Mug", Price: 50, PurchaseDate: time.Now()},
	}

	// act
	purchaseRepo.On("GetByUser", mock.Anything, mock.Anything, userID).Return(purchases, nil)

	service := services.NewMerchService(merchRepo, purchaseRepo, userRepo, db)

	result, err := service.GetPurchasesByUser(context.Background(), userID)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, purchases, result)
	purchaseRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestMerchService_GetPurchasesByUser_BeginTxError(t *testing.T) {
	// arrange
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	merchRepo := new(mocks.MockMerchRepository)
	purchaseRepo := new(mocks.MockPurchaseRepository)
	userRepo := new(mocks.MockUserRepository)

	beginErr := errors.New("begin failed")
	mockDB.ExpectBegin().WillReturnError(beginErr)

	userID := int64(1)

	service := services.NewMerchService(merchRepo, purchaseRepo, userRepo, db)

	// act
	_, err = service.GetPurchasesByUser(context.Background(), userID)

	// assert
	assert.Error(t, err)
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		assert.Equal(t, beginErr, unwrapped)
	}
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestMerchService_GetPurchasesByUser_RepositoryError(t *testing.T) {
	// arrange
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	merchRepo := new(mocks.MockMerchRepository)
	purchaseRepo := new(mocks.MockPurchaseRepository)
	userRepo := new(mocks.MockUserRepository)

	mockDB.ExpectBegin()

	userID := int64(1)
	repoErr := errors.New("repository error")

	// act
	purchaseRepo.On("GetByUser", mock.Anything, mock.Anything, userID).Return(nil, repoErr)

	service := services.NewMerchService(merchRepo, purchaseRepo, userRepo, db)

	_, err = service.GetPurchasesByUser(context.Background(), userID)

	// assert
	assert.Error(t, err)
	assert.Equal(t, repoErr, err)
	purchaseRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestMerchService_GetAllAvailableMerch_Success(t *testing.T) {
	// arrange
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	merchRepo := new(mocks.MockMerchRepository)
	purchaseRepo := new(mocks.MockPurchaseRepository)
	userRepo := new(mocks.MockUserRepository)

	merch := []*domain.Merch{
		{ID: 1, Name: "T-Shirt", Price: 100},
		{ID: 2, Name: "Mug", Price: 50},
	}

	// act
	merchRepo.On("GetAllAvailableMerch", mock.Anything).Return(merch, nil)

	service := services.NewMerchService(merchRepo, purchaseRepo, userRepo, db)

	result, err := service.GetAllAvailableMerch(context.Background())

	// assert
	assert.NoError(t, err)
	assert.Equal(t, merch, result)
	merchRepo.AssertExpectations(t)
}

func TestMerchService_GetAllAvailableMerch_RepositoryError(t *testing.T) {
	// arrange
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	merchRepo := new(mocks.MockMerchRepository)
	purchaseRepo := new(mocks.MockPurchaseRepository)
	userRepo := new(mocks.MockUserRepository)

	repoErr := errors.New("repository error")

	// act
	merchRepo.On("GetAllAvailableMerch", mock.Anything).Return(nil, repoErr)

	service := services.NewMerchService(merchRepo, purchaseRepo, userRepo, db)

	_, err = service.GetAllAvailableMerch(context.Background())

	// assert
	assert.Error(t, err)
	assert.Equal(t, repoErr, err)
	merchRepo.AssertExpectations(t)
}

func BenchmarkPurchaseItem(b *testing.B) {
	userID := int64(1)
	itemName := "T-Shirt"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			db, mockDB, err := sqlmock.New()
			if err != nil {
				b.Fatal(err)
			}
			defer db.Close()

			mockDB.ExpectBegin()
			mockDB.ExpectCommit()

			merchRepo := new(mocks.MockMerchRepository)
			purchaseRepo := new(mocks.MockPurchaseRepository)
			userRepo := new(mocks.MockUserRepository)

			merchRepo.On("FindByName", mock.Anything, itemName).Return(&domain.Merch{
				ID:    1,
				Name:  itemName,
				Price: 100,
			}, nil)
			userRepo.On("FindByID", mock.Anything, userID).Return(&domain.User{
				ID:    userID,
				Coins: 200,
			}, nil)
			userRepo.On("Update", mock.Anything, mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
				return u.ID == userID && u.Coins == 100
			})).Return(nil)
			purchaseRepo.On("Create", mock.Anything, mock.Anything, mock.MatchedBy(func(p *domain.Purchase) bool {
				return p.UserID == userID && p.Item == itemName && p.Price == 100
			})).Return(nil)

			service := services.NewMerchService(merchRepo, purchaseRepo, userRepo, db)
			err = service.PurchaseItem(context.Background(), userID, itemName)
			if err != nil {
				b.Error(err)
			}

			if err := mockDB.ExpectationsWereMet(); err != nil {
				b.Error(err)
			}
		}
	})
}
