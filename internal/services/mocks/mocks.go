package mocks

import (
	"avito-backend-intern-winter25/internal/models/domain"
	"avito-backend-intern-winter25/internal/services/jwt"
	"avito-backend-intern-winter25/internal/storage"
	"context"
	"database/sql"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"
	"time"
)

type MockUserRepository struct {
	mock.Mock
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{}
}

func (m *MockUserRepository) BeginTx(ctx context.Context) (storage.Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(storage.Tx), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, tx storage.Tx, user *domain.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, tx storage.Tx, user *domain.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByIDForUpdate(ctx context.Context, tx storage.Tx, id int64) (*domain.User, error) {
	args := m.Called(ctx, tx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

type MockJWT struct {
	mock.Mock
}

func NewMockJWT() *MockJWT {
	return &MockJWT{}
}

func (m *MockJWT) GenerateToken(userID int64, username string) (string, error) {
	args := m.Called(userID, username)
	return args.String(0), args.Error(1)
}

func (m *MockJWT) ValidateToken(tokenString string) (*jwt.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Claims), args.Error(1)
}

type MockTx struct {
	mock.Mock
}

func (m *MockTx) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTx) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	callArgs := m.Called(ctx, query, args)
	res, _ := callArgs.Get(0).(sql.Result)
	return res, callArgs.Error(1)
}

func (m *MockTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	callArgs := m.Called(ctx, query, args)
	row, _ := callArgs.Get(0).(*sql.Row)
	return row
}

type MockTransactionRepository struct {
	mock.Mock
}

func NewMockTransactionRepository() *MockTransactionRepository {
	return &MockTransactionRepository{}
}

func (m *MockTransactionRepository) Create(ctx context.Context, tx storage.Tx, transaction *domain.CoinTransaction) error {
	args := m.Called(ctx, tx, transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetSentTransactions(ctx context.Context, userID int64) ([]*domain.CoinTransaction, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CoinTransaction), args.Error(1)
}

func (m *MockTransactionRepository) GetReceivedTransactions(ctx context.Context, userID int64) ([]*domain.CoinTransaction, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CoinTransaction), args.Error(1)
}

type MockMerchRepository struct {
	mock.Mock
}

func NewMockMerchRepository() *MockMerchRepository {
	return &MockMerchRepository{}
}

func (m *MockMerchRepository) GetAllAvailableMerch(ctx context.Context) ([]*domain.Merch, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Merch), args.Error(1)
}

func (m *MockMerchRepository) FindByName(ctx context.Context, name string) (*domain.Merch, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Merch), args.Error(1)
}

type MockPurchaseRepository struct {
	mock.Mock
}

func NewMockPurchaseRepository() *MockPurchaseRepository {
	return &MockPurchaseRepository{}
}

func (m *MockPurchaseRepository) Create(ctx context.Context, tx *sql.Tx, purchase *domain.Purchase) error {
	args := m.Called(ctx, tx, purchase)
	return args.Error(0)
}

func (m *MockPurchaseRepository) GetByUser(ctx context.Context, tx *sql.Tx, userID int64) ([]*domain.Purchase, error) {
	args := m.Called(ctx, tx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Purchase), args.Error(1)
}

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	if cmd, ok := args.Get(0).(*redis.StringCmd); ok {
		return cmd
	}
	return redis.NewStringResult("", redis.Nil)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	if cmd, ok := args.Get(0).(*redis.StatusCmd); ok {
		return cmd
	}
	return redis.NewStatusResult("OK", nil)
}
