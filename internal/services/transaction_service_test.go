package services

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"avito-backend-intern-winter25/internal/services/mocks"
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestTransferCoins_BeginTxError(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	expectedErr := errors.New("begin tx error")
	mockDB.ExpectBegin().WillReturnError(expectedErr)

	userRepo := new(mocks.MockUserRepository)
	transactionRepo := new(mocks.MockTransactionRepository)

	service := NewTransactionService(db, userRepo, transactionRepo)
	err = service.TransferCoins(context.Background(), 1, 2, 100)
	assert.ErrorIs(t, err, expectedErr)

	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestTransferCoins_FindFromUserError(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockDB.ExpectBegin()

	expectedErr := errors.New("find from user error")
	userRepo := new(mocks.MockUserRepository)
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(1)).
		Return(nil, expectedErr)

	transactionRepo := new(mocks.MockTransactionRepository)
	mockDB.ExpectRollback()

	service := NewTransactionService(db, userRepo, transactionRepo)
	err = service.TransferCoins(context.Background(), 1, 2, 100)
	assert.ErrorIs(t, err, expectedErr)

	userRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestTransferCoins_FindToUserError(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockDB.ExpectBegin()

	fromUser := &domain.User{ID: 1, Coins: 200}
	userRepo := new(mocks.MockUserRepository)
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(1)).
		Return(fromUser, nil)
	expectedErr := errors.New("find to user error")
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(2)).
		Return(nil, expectedErr)

	transactionRepo := new(mocks.MockTransactionRepository)
	mockDB.ExpectRollback()

	service := NewTransactionService(db, userRepo, transactionRepo)
	err = service.TransferCoins(context.Background(), 1, 2, 100)
	assert.ErrorIs(t, err, expectedErr)

	userRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestTransferCoins_InsufficientFunds(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockDB.ExpectBegin()

	fromUser := &domain.User{ID: 1, Coins: 50}
	toUser := &domain.User{ID: 2, Coins: 300}
	userRepo := new(mocks.MockUserRepository)
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(1)).
		Return(fromUser, nil)
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(2)).
		Return(toUser, nil)

	transactionRepo := new(mocks.MockTransactionRepository)
	mockDB.ExpectRollback()

	service := NewTransactionService(db, userRepo, transactionRepo)
	err = service.TransferCoins(context.Background(), 1, 2, 100)
	assert.ErrorIs(t, err, ErrLackOfFundsOnAccount)

	userRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestTransferCoins_UpdateFromUserError(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockDB.ExpectBegin()

	fromUser := &domain.User{ID: 1, Coins: 200}
	toUser := &domain.User{ID: 2, Coins: 300}
	userRepo := new(mocks.MockUserRepository)
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(1)).
		Return(fromUser, nil)
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(2)).
		Return(toUser, nil)

	expectedErr := errors.New("update from user error")
	userRepo.
		On("Update", mock.Anything, mock.Anything, fromUser).
		Return(expectedErr)

	transactionRepo := new(mocks.MockTransactionRepository)
	mockDB.ExpectRollback()

	service := NewTransactionService(db, userRepo, transactionRepo)
	err = service.TransferCoins(context.Background(), 1, 2, 100)
	assert.ErrorIs(t, err, expectedErr)

	userRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestTransferCoins_UpdateToUserError(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockDB.ExpectBegin()

	fromUser := &domain.User{ID: 1, Coins: 200}
	toUser := &domain.User{ID: 2, Coins: 300}
	userRepo := new(mocks.MockUserRepository)
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(1)).
		Return(fromUser, nil)
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(2)).
		Return(toUser, nil)

	userRepo.
		On("Update", mock.Anything, mock.Anything, fromUser).
		Return(nil)
	expectedErr := errors.New("update to user error")
	userRepo.
		On("Update", mock.Anything, mock.Anything, toUser).
		Return(expectedErr)

	transactionRepo := new(mocks.MockTransactionRepository)
	mockDB.ExpectRollback()

	service := NewTransactionService(db, userRepo, transactionRepo)
	err = service.TransferCoins(context.Background(), 1, 2, 100)
	assert.ErrorIs(t, err, expectedErr)

	userRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestTransferCoins_CreateTransactionError(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockDB.ExpectBegin()

	fromUser := &domain.User{ID: 1, Coins: 200}
	toUser := &domain.User{ID: 2, Coins: 300}
	userRepo := new(mocks.MockUserRepository)
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(1)).
		Return(fromUser, nil)
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(2)).
		Return(toUser, nil)
	userRepo.
		On("Update", mock.Anything, mock.Anything, fromUser).
		Return(nil)
	userRepo.
		On("Update", mock.Anything, mock.Anything, toUser).
		Return(nil)

	expectedErr := errors.New("create transaction error")
	transactionRepo := new(mocks.MockTransactionRepository)
	transactionRepo.
		On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*domain.CoinTransaction")).
		Return(expectedErr)

	mockDB.ExpectRollback()

	service := NewTransactionService(db, userRepo, transactionRepo)
	err = service.TransferCoins(context.Background(), 1, 2, 100)
	assert.ErrorIs(t, err, expectedErr)

	userRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestTransferCoins_CommitError(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockDB.ExpectBegin()

	fromUser := &domain.User{ID: 1, Coins: 200}
	toUser := &domain.User{ID: 2, Coins: 300}
	userRepo := new(mocks.MockUserRepository)
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(1)).
		Return(fromUser, nil)
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(2)).
		Return(toUser, nil)

	userRepo.
		On("Update", mock.Anything, mock.Anything, fromUser).
		Return(nil)
	userRepo.
		On("Update", mock.Anything, mock.Anything, toUser).
		Return(nil)

	transactionRepo := new(mocks.MockTransactionRepository)
	transactionRepo.
		On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*domain.CoinTransaction")).
		Return(nil)

	commitError := errors.New("commit error")
	mockDB.ExpectCommit().WillReturnError(commitError)

	service := NewTransactionService(db, userRepo, transactionRepo)
	err = service.TransferCoins(context.Background(), 1, 2, 100)
	assert.ErrorIs(t, err, commitError)

	userRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)

	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestTransferCoins_Success(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockDB.ExpectBegin()

	fromUser := &domain.User{ID: 1, Coins: 200}
	toUser := &domain.User{ID: 2, Coins: 300}
	userRepo := new(mocks.MockUserRepository)
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(1)).
		Return(fromUser, nil)
	userRepo.
		On("FindByIDForUpdate", mock.Anything, mock.Anything, int64(2)).
		Return(toUser, nil)

	userRepo.
		On("Update", mock.Anything, mock.Anything, fromUser).
		Return(nil)
	userRepo.
		On("Update", mock.Anything, mock.Anything, toUser).
		Return(nil)

	transactionRepo := new(mocks.MockTransactionRepository)
	transactionRepo.
		On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*domain.CoinTransaction")).
		Return(nil)

	mockDB.ExpectCommit()

	service := NewTransactionService(db, userRepo, transactionRepo)
	err = service.TransferCoins(context.Background(), 1, 2, 100)
	assert.NoError(t, err)

	userRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestGetSentTransactions_Success(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	expectedTransactions := []*domain.CoinTransaction{
		{FromUserID: userID, ToUserID: 2, Amount: 100, CreatedAt: time.Now()},
	}
	transactionRepo := new(mocks.MockTransactionRepository)
	transactionRepo.
		On("GetSentTransactions", ctx, userID).
		Return(expectedTransactions, nil)

	service := NewTransactionService(nil, nil, transactionRepo)
	transactions, err := service.GetSentTransactions(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedTransactions, transactions)
	transactionRepo.AssertExpectations(t)
}

func TestGetSentTransactions_Error(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	expectedErr := errors.New("get sent error")
	transactionRepo := new(mocks.MockTransactionRepository)
	transactionRepo.
		On("GetSentTransactions", ctx, userID).
		Return(nil, expectedErr)

	service := NewTransactionService(nil, nil, transactionRepo)
	transactions, err := service.GetSentTransactions(ctx, userID)
	assert.ErrorIs(t, err, expectedErr)
	assert.Nil(t, transactions)
	transactionRepo.AssertExpectations(t)
}

func TestGetReceivedTransactions_Success(t *testing.T) {
	ctx := context.Background()
	userID := int64(2)
	expectedTransactions := []*domain.CoinTransaction{
		{FromUserID: 1, ToUserID: userID, Amount: 50, CreatedAt: time.Now()},
	}
	transactionRepo := new(mocks.MockTransactionRepository)
	transactionRepo.
		On("GetReceivedTransactions", ctx, userID).
		Return(expectedTransactions, nil)

	service := NewTransactionService(nil, nil, transactionRepo)
	transactions, err := service.GetReceivedTransactions(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedTransactions, transactions)
	transactionRepo.AssertExpectations(t)
}

func TestGetReceivedTransactions_Error(t *testing.T) {
	ctx := context.Background()
	userID := int64(2)
	expectedErr := errors.New("get received error")
	transactionRepo := new(mocks.MockTransactionRepository)
	transactionRepo.
		On("GetReceivedTransactions", ctx, userID).
		Return(nil, expectedErr)

	service := NewTransactionService(nil, nil, transactionRepo)
	transactions, err := service.GetReceivedTransactions(ctx, userID)
	assert.ErrorIs(t, err, expectedErr)
	assert.Nil(t, transactions)
	transactionRepo.AssertExpectations(t)
}
