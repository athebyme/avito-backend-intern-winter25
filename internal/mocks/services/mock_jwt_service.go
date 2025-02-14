package mocks_services

import (
	"avito-backend-intern-winter25/internal/services/jwt"
	"github.com/stretchr/testify/mock"
)

type MockJWTService struct {
	mock.Mock
}

func NewMockJWTService() *MockJWTService {
	return &MockJWTService{}
}

func (m *MockJWTService) GenerateToken(userID int64, username string) (string, error) {
	return "test-token", nil
}

func (m *MockJWTService) ValidateToken(tokenString string) (*jwt.Claims, error) {
	args := m.Called(tokenString)
	return args.Get(0).(*jwt.Claims), args.Error(1)
}
