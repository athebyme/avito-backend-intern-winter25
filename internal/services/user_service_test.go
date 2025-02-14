package services

import (
	mocks_tx "avito-backend-intern-winter25/internal/mocks/mock_tx"
	mocks_repo "avito-backend-intern-winter25/internal/mocks/repo"
	mocks_services "avito-backend-intern-winter25/internal/mocks/services"
	"avito-backend-intern-winter25/internal/models/domain"
	"avito-backend-intern-winter25/internal/storage"
	"context"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestUserService_Login_InvalidPassword(t *testing.T) {
	ctx := context.Background()
	username := "existinguser"
	password := "wrongpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	existingUser := &domain.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Coins:        1500,
	}

	mockUserRepo := new(mocks_repo.MockUserRepository)
	mockJWTService := new(mocks_services.MockJWTService)
	mockTx := new(mocks_tx.MockTx)

	service := NewUserService(mockUserRepo, mockJWTService)

	mockUserRepo.On("BeginTx", ctx).Return(mockTx, nil)
	mockUserRepo.On("FindByUsername", ctx, username).Return(existingUser, nil)
	mockTx.On("Rollback").Return(nil)

	_, err := service.Login(ctx, username, password)

	assert.ErrorIs(t, err, ErrInvalidPassword)
	mockTx.AssertExpectations(t)
}

func TestUserService_GenerateToken(t *testing.T) {
	user := &domain.User{
		ID:       1,
		Username: "testuser",
	}

	mockUserRepo := new(mocks_repo.MockUserRepository)
	mockJWTService := new(mocks_services.MockJWTService)

	service := NewUserService(mockUserRepo, mockJWTService)

	expectedToken := "test.token"
	mockJWTService.On("GenerateToken", user.ID, user.Username).Return(expectedToken, nil)

	token, err := service.GenerateToken(user)

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	mockJWTService.AssertExpectations(t)
}

func TestUserService_GetUserByID_Success(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	expectedUser := &domain.User{ID: userID}

	mockUserRepo := new(mocks_repo.MockUserRepository)
	service := NewUserService(mockUserRepo, nil)

	mockUserRepo.On("FindByID", ctx, userID).Return(expectedUser, nil)

	user, err := service.GetUserByID(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)

	mockUserRepo := new(mocks_repo.MockUserRepository)
	service := NewUserService(mockUserRepo, nil)

	mockUserRepo.On("FindByID", ctx, userID).Return(nil, storage.ErrUserNotFound)

	_, err := service.GetUserByID(ctx, userID)

	assert.ErrorIs(t, err, ErrUserNotFound)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_UpdateUserCoins(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	newCoins := 2000

	mockUserRepo := new(mocks_repo.MockUserRepository)
	mockTx := new(mocks_tx.MockTx)
	service := NewUserService(mockUserRepo, nil)

	existingUser := &domain.User{ID: userID, Coins: 1000}

	mockUserRepo.On("BeginTx", ctx).Return(mockTx, nil)
	mockUserRepo.On("FindByID", ctx, userID).Return(existingUser, nil)
	mockUserRepo.On("Update", ctx, mockTx, &domain.User{ID: userID, Coins: newCoins}).Return(nil)
	mockTx.On("Commit").Return(nil)

	err := service.UpdateUserCoins(ctx, userID, newCoins)

	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestUserService_ValidateUserBalance_Sufficient(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	amount := 500

	mockUserRepo := new(mocks_repo.MockUserRepository)
	service := NewUserService(mockUserRepo, nil)

	existingUser := &domain.User{ID: userID, Coins: 1000}
	mockUserRepo.On("FindByID", ctx, userID).Return(existingUser, nil)

	valid, err := service.ValidateUserBalance(ctx, userID, amount)

	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestUserService_ValidateUserBalance_Insufficient(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	amount := 1500

	mockUserRepo := new(mocks_repo.MockUserRepository)
	service := NewUserService(mockUserRepo, nil)

	existingUser := &domain.User{ID: userID, Coins: 1000}
	mockUserRepo.On("FindByID", ctx, userID).Return(existingUser, nil)

	valid, err := service.ValidateUserBalance(ctx, userID, amount)

	assert.NoError(t, err)
	assert.False(t, valid)
}
