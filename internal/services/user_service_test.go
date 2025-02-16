package services

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"avito-backend-intern-winter25/internal/services/jwt"
	"avito-backend-intern-winter25/internal/services/mocks"
	"avito-backend-intern-winter25/internal/storage"
	"context"
	"database/sql"
	"errors"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

func TestNewUserService(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)
	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	assert.NotNil(t, service)
}

func TestLogin_NewUser(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)
	mockTx := new(mocks.MockTx)

	ctx := context.Background()
	username := "newuser"
	password := "password123"

	mockUserRepo.On("BeginTx", ctx).Return(mockTx, nil)
	mockUserRepo.On("FindByUsername", ctx, username).Return(nil, storage.ErrUserNotFound)
	mockUserRepo.On("Create", ctx, mockTx, mock.AnythingOfType("*domain.User")).Return(nil)
	mockTx.On("Commit").Return(nil)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	user, err := service.Login(ctx, username, password)

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, 1000, user.Coins)

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestLogin_ExistingUser_ValidPassword(t *testing.T) {
	// arrange
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)
	ctx := context.Background()
	username := "existinguser"
	password := "password123"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	fixedTime := time.Date(2025, 2, 16, 15, 0, 0, 0, time.UTC)
	existingUser := &domain.User{
		ID:           1,
		Username:     username,
		PasswordHash: string(passwordHash),
		Coins:        500,
		CreatedAt:    fixedTime,
	}

	redisClient, redisMock := redismock.NewClientMock()

	redisMock.ExpectGet("user:" + username).RedisNil()

	redisMock.Regexp().ExpectSet("user:"+username, ".*", 30*time.Minute).SetVal("OK")

	mockUserRepo.On("FindByUsername", ctx, username).Return(existingUser, nil)

	// act
	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	user, err := service.Login(ctx, username, password)

	// assert
	require.NoError(t, err)
	assert.Equal(t, existingUser.ID, user.ID)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, 500, user.Coins)

	if err := redisMock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet redis expectations: %s", err)
	}
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_ExistingUser_InvalidPassword(t *testing.T) {
	// arrange
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	ctx := context.Background()
	username := "existinguser"
	correctPassword := "password123"
	wrongPassword := "wrongpassword"

	fixedTime := time.Date(2025, 2, 16, 15, 0, 0, 0, time.UTC)
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	existingUser := &domain.User{
		ID:           1,
		Username:     username,
		PasswordHash: string(passwordHash),
		Coins:        500,
		CreatedAt:    fixedTime,
	}

	redisClient, redisMock := redismock.NewClientMock()
	redisMock.ExpectGet("user:" + username).RedisNil()

	mockUserRepo.On("FindByUsername", ctx, username).Return(existingUser, nil)

	// act
	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	user, err := service.Login(ctx, username, wrongPassword)

	// assert
	assert.Nil(t, user)
	assert.ErrorIs(t, err, ErrInvalidPassword)

	if err := redisMock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet redis expectations: %s", err)
	}
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_BeginTxError(t *testing.T) {
	// arrange
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	ctx := context.Background()
	username := "username"
	password := "password"
	expectedErr := errors.New("db connection error")

	redisClient, redisMock := redismock.NewClientMock()

	redisMock.ExpectGet("user:" + username).RedisNil()

	mockUserRepo.On("FindByUsername", ctx, username).Return(nil, storage.ErrUserNotFound)
	mockUserRepo.On("BeginTx", ctx).Return(nil, expectedErr)

	// act
	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	user, err := service.Login(ctx, username, password)

	// assert
	assert.Nil(t, user)
	assert.ErrorIs(t, err, expectedErr)

	if err := redisMock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet redis expectations: %s", err)
	}
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_FindByUsernameError(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)
	mockTx := new(mocks.MockTx)

	ctx := context.Background()
	username := "username"
	expectedErr := errors.New("db query error")

	mockUserRepo.On("BeginTx", ctx).Return(mockTx, nil)
	mockUserRepo.On("FindByUsername", ctx, username).Return(nil, expectedErr)
	mockTx.On("Rollback").Return(nil)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	user, err := service.Login(ctx, username, "password")

	assert.Nil(t, user)
	assert.ErrorIs(t, err, expectedErr)

	mockUserRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestLogin_CreateError(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)
	mockTx := new(mocks.MockTx)

	ctx := context.Background()
	username := "newuser"
	expectedErr := errors.New("create user error")

	mockUserRepo.On("BeginTx", ctx).Return(mockTx, nil)
	mockUserRepo.On("FindByUsername", ctx, username).Return(nil, storage.ErrUserNotFound)
	mockUserRepo.On("Create", ctx, mockTx, mock.AnythingOfType("*domain.User")).Return(expectedErr)
	mockTx.On("Rollback").Return(nil)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	user, err := service.Login(ctx, username, "password")

	assert.Nil(t, user)
	assert.ErrorIs(t, err, expectedErr)

	mockUserRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestLogin_CommitError(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)
	mockTx := new(mocks.MockTx)

	ctx := context.Background()
	username := "newuser"
	expectedErr := errors.New("commit error")

	mockUserRepo.On("BeginTx", ctx).Return(mockTx, nil)
	mockUserRepo.On("FindByUsername", ctx, username).Return(nil, storage.ErrUserNotFound)
	mockUserRepo.On("Create", ctx, mockTx, mock.AnythingOfType("*domain.User")).Return(nil)
	mockTx.On("Commit").Return(expectedErr)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	user, err := service.Login(ctx, username, "password")

	assert.Nil(t, user)
	assert.ErrorIs(t, err, expectedErr)

	mockUserRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestGenerateToken(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	user := &domain.User{
		ID:       123,
		Username: "testuser",
	}
	expectedToken := "test.jwt.token"

	mockJWT.On("GenerateToken", int64(123), "testuser").Return(expectedToken, nil)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	token, err := service.GenerateToken(user)

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)

	mockJWT.AssertExpectations(t)
}

func TestGenerateToken_Error(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	user := &domain.User{
		ID:       123,
		Username: "testuser",
	}
	expectedErr := errors.New("token generation error")

	mockJWT.On("GenerateToken", int64(123), "testuser").Return("", expectedErr)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	token, err := service.GenerateToken(user)

	assert.Equal(t, "", token)
	assert.ErrorIs(t, err, expectedErr)

	mockJWT.AssertExpectations(t)
}

func TestGetUserByID(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	ctx := context.Background()
	userId := int64(123)
	expectedUser := &domain.User{
		ID:       userId,
		Username: "testuser",
		Coins:    500,
	}

	mockUserRepo.On("FindByID", ctx, userId).Return(expectedUser, nil)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	user, err := service.GetUserByID(ctx, userId)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)

	mockUserRepo.AssertExpectations(t)
}

func TestGetUserByID_NotFound(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	ctx := context.Background()
	userId := int64(999)

	mockUserRepo.On("FindByID", ctx, userId).Return(nil, storage.ErrUserNotFound)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	user, err := service.GetUserByID(ctx, userId)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, ErrUserNotFound)

	mockUserRepo.AssertExpectations(t)
}

func TestGetUserByID_OtherError(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	ctx := context.Background()
	userId := int64(123)
	expectedErr := errors.New("database error")

	mockUserRepo.On("FindByID", ctx, userId).Return(nil, expectedErr)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	user, err := service.GetUserByID(ctx, userId)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, expectedErr)

	mockUserRepo.AssertExpectations(t)
}

func TestGetUserByUsername(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	ctx := context.Background()
	username := "testuser"
	expectedUser := &domain.User{
		ID:       123,
		Username: username,
		Coins:    500,
	}

	mockUserRepo.On("FindByUsername", ctx, username).Return(expectedUser, nil)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	user, err := service.GetUserByUsername(ctx, username)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)

	mockUserRepo.AssertExpectations(t)
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	ctx := context.Background()
	username := "nonexistentuser"

	mockUserRepo.On("FindByUsername", ctx, username).Return(nil, storage.ErrUserNotFound)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	user, err := service.GetUserByUsername(ctx, username)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, ErrUserNotFound)

	mockUserRepo.AssertExpectations(t)
}

func TestGetUserByUsername_OtherError(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	ctx := context.Background()
	username := "testuser"
	expectedErr := errors.New("database error")

	mockUserRepo.On("FindByUsername", ctx, username).Return(nil, expectedErr)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	user, err := service.GetUserByUsername(ctx, username)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, expectedErr)

	mockUserRepo.AssertExpectations(t)
}

func TestUpdateUserCoins(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)
	mockTx := new(mocks.MockTx)

	ctx := context.Background()
	userId := int64(123)
	newCoins := 750

	existingUser := &domain.User{
		ID:       userId,
		Username: "testuser",
		Coins:    500,
	}

	mockUserRepo.On("BeginTx", ctx).Return(mockTx, nil)
	mockUserRepo.On("FindByID", ctx, userId).Return(existingUser, nil)
	mockUserRepo.On("Update", ctx, mockTx, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == userId && u.Coins == newCoins
	})).Return(nil)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	err := service.UpdateUserCoins(ctx, userId, newCoins)

	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
}

func TestUpdateUserCoins_BeginTxError(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	ctx := context.Background()
	userId := int64(123)
	newCoins := 750
	expectedErr := errors.New("begin transaction error")

	mockUserRepo.On("BeginTx", ctx).Return(nil, expectedErr)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	err := service.UpdateUserCoins(ctx, userId, newCoins)

	assert.ErrorIs(t, err, expectedErr)

	mockUserRepo.AssertExpectations(t)
}

func TestUpdateUserCoins_FindByIDError(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)
	mockTx := new(mocks.MockTx)

	ctx := context.Background()
	userId := int64(123)
	newCoins := 750
	expectedErr := errors.New("find user error")

	mockUserRepo.On("BeginTx", ctx).Return(mockTx, nil)
	mockUserRepo.On("FindByID", ctx, userId).Return(nil, expectedErr)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	err := service.UpdateUserCoins(ctx, userId, newCoins)

	assert.ErrorIs(t, err, expectedErr)

	mockUserRepo.AssertExpectations(t)
}

func TestUpdateUserCoins_UpdateError(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)
	mockTx := new(mocks.MockTx)

	ctx := context.Background()
	userId := int64(123)
	newCoins := 750
	expectedErr := errors.New("update user error")

	existingUser := &domain.User{
		ID:       userId,
		Username: "testuser",
		Coins:    500,
	}

	mockUserRepo.On("BeginTx", ctx).Return(mockTx, nil)
	mockUserRepo.On("FindByID", ctx, userId).Return(existingUser, nil)
	mockUserRepo.On("Update", ctx, mockTx, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == userId && u.Coins == newCoins
	})).Return(expectedErr)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	err := service.UpdateUserCoins(ctx, userId, newCoins)

	assert.ErrorIs(t, err, expectedErr)

	mockUserRepo.AssertExpectations(t)
}

func TestValidateUserBalance_Sufficient(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	ctx := context.Background()
	userId := int64(123)
	amount := 300

	user := &domain.User{
		ID:       userId,
		Username: "testuser",
		Coins:    500,
	}

	mockUserRepo.On("FindByID", ctx, userId).Return(user, nil)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	sufficient, err := service.ValidateUserBalance(ctx, userId, amount)

	assert.NoError(t, err)
	assert.True(t, sufficient)

	mockUserRepo.AssertExpectations(t)
}

func TestValidateUserBalance_Insufficient(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	ctx := context.Background()
	userId := int64(123)
	amount := 600

	user := &domain.User{
		ID:       userId,
		Username: "testuser",
		Coins:    500,
	}

	mockUserRepo.On("FindByID", ctx, userId).Return(user, nil)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	sufficient, err := service.ValidateUserBalance(ctx, userId, amount)

	assert.NoError(t, err)
	assert.False(t, sufficient)

	mockUserRepo.AssertExpectations(t)
}

func TestValidateUserBalance_Error(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	ctx := context.Background()
	userId := int64(123)
	amount := 300
	expectedErr := errors.New("user not found")

	mockUserRepo.On("FindByID", ctx, userId).Return(nil, expectedErr)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	sufficient, err := service.ValidateUserBalance(ctx, userId, amount)

	assert.ErrorIs(t, err, expectedErr)
	assert.False(t, sufficient)

	mockUserRepo.AssertExpectations(t)
}

func TestGetUserBalance(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	ctx := context.Background()
	userId := int64(123)
	expectedCoins := 500

	user := &domain.User{
		ID:       userId,
		Username: "testuser",
		Coins:    expectedCoins,
	}

	mockUserRepo.On("FindByID", ctx, userId).Return(user, nil)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	coins, err := service.GetUserBalance(ctx, userId)

	assert.NoError(t, err)
	assert.Equal(t, expectedCoins, coins)

	mockUserRepo.AssertExpectations(t)
}

func TestGetUserBalance_Error(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)

	ctx := context.Background()
	userId := int64(123)
	expectedErr := errors.New("user not found")

	mockUserRepo.On("FindByID", ctx, userId).Return(nil, expectedErr)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)
	coins, err := service.GetUserBalance(ctx, userId)

	assert.Equal(t, 0, coins)
	assert.ErrorIs(t, err, expectedErr)

	mockUserRepo.AssertExpectations(t)
}

func TestLogin_PanicInTransaction(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWT := new(mocks.MockJWT)
	mockTx := new(mocks.MockTx)

	ctx := context.Background()
	username := "username"

	mockUserRepo.On("BeginTx", ctx).Return(mockTx, nil)
	mockUserRepo.On("FindByUsername", ctx, username).Run(func(args mock.Arguments) {
		panic("unexpected panic")
	})
	mockTx.On("Rollback").Return(nil)

	redisClient, _ := redismock.NewClientMock()

	service := NewUserService(mockUserRepo, mockJWT, redisClient)

	assert.Panics(t, func() {
		_, _ = service.Login(ctx, username, "password")
	})

	mockTx.AssertExpectations(t)
}

type dummyTx struct{}

func (d dummyTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	panic("implement me")
}

func (d dummyTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	panic("implement me")
}

func (d dummyTx) Commit() error   { return nil }
func (d dummyTx) Rollback() error { return nil }

type dummyUserRepository struct {
	hashedPassword string
}

func (d *dummyUserRepository) BeginTx(ctx context.Context) (storage.Tx, error) {
	return dummyTx{}, nil
}

func (d *dummyUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	return &domain.User{
		ID:           1,
		Username:     username,
		PasswordHash: d.hashedPassword,
		Coins:        1000,
	}, nil
}

func (d *dummyUserRepository) Create(ctx context.Context, tx storage.Tx, user *domain.User) error {
	return nil
}
func (d *dummyUserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	return nil, sql.ErrNoRows
}
func (d *dummyUserRepository) Update(ctx context.Context, tx storage.Tx, user *domain.User) error {
	return nil
}
func (d *dummyUserRepository) FindByIDForUpdate(ctx context.Context, tx storage.Tx, id int64) (*domain.User, error) {
	return nil, sql.ErrNoRows
}

type dummyJWT struct{}

func (d dummyJWT) GenerateToken(userID int64, username string) (string, error) {
	return "dummyToken", nil
}
func (d dummyJWT) ValidateToken(tokenString string) (*jwt.Claims, error) {
	return &jwt.Claims{UserID: 0, Username: ""}, nil
}

func BenchmarkUserService_Login(b *testing.B) {
	ctx := context.Background()
	password := "password"

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		b.Fatal(err)
	}

	repo := &dummyUserRepository{hashedPassword: string(hashed)}
	jwtSvc := dummyJWT{}
	redisClient, _ := redismock.NewClientMock()

	userSvc := NewUserService(repo, jwtSvc, redisClient)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			user, err := userSvc.Login(ctx, "testuser", password)
			if err != nil {
				b.Fatal(err)
			}
			if user == nil {
				b.Fatal("expected user, got nil")
			}
		}
	})
}
