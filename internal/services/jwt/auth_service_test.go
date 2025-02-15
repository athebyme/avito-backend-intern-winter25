package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewService(t *testing.T) {
	secretKey := "test-secret"
	tokenLifetime := 24 * time.Hour

	service := NewService(secretKey, tokenLifetime)
	assert.NotNil(t, service)
}

func TestGenerateToken(t *testing.T) {
	secretKey := "test-secret"
	tokenLifetime := 24 * time.Hour
	service := NewService(secretKey, tokenLifetime)

	userID := int64(123)
	username := "testuser"

	token, err := service.GenerateToken(userID, username)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)

	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)
	assert.NotNil(t, claims.ExpiresAt)

	now := time.Now()
	assert.True(t, claims.IssuedAt.Time.Before(now.Add(1*time.Minute)))
	assert.True(t, claims.IssuedAt.Time.After(now.Add(-1*time.Minute)))

	assert.True(t, claims.ExpiresAt.Time.Before(now.Add(tokenLifetime).Add(1*time.Minute)))
	assert.True(t, claims.ExpiresAt.Time.After(now.Add(tokenLifetime).Add(-1*time.Minute)))
}

func TestValidateToken_Valid(t *testing.T) {
	secretKey := "test-secret"
	tokenLifetime := 24 * time.Hour
	service := NewService(secretKey, tokenLifetime)

	userID := int64(123)
	username := "testuser"

	token, err := service.GenerateToken(userID, username)
	require.NoError(t, err)

	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
}

func TestValidateToken_Expired(t *testing.T) {
	secretKey := "test-secret"
	tokenLifetime := -1 * time.Hour
	service := NewService(secretKey, tokenLifetime)

	token, err := service.GenerateToken(123, "testuser")
	require.NoError(t, err)

	claims, err := service.ValidateToken(token)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	service1 := NewService("secret1", 24*time.Hour)
	service2 := NewService("secret2", 24*time.Hour)

	token, err := service1.GenerateToken(123, "testuser")
	require.NoError(t, err)

	claims, err := service2.ValidateToken(token)
	assert.Nil(t, claims)
	assert.Error(t, err)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	service := NewService("test-secret", 24*time.Hour)

	claims, err := service.ValidateToken("invalid.token.string")
	assert.Nil(t, claims)
	assert.Error(t, err)
}

func TestValidateToken_EmptyToken(t *testing.T) {
	service := NewService("test-secret", 24*time.Hour)

	claims, err := service.ValidateToken("")
	assert.Nil(t, claims)
	assert.Error(t, err)
}

func TestValidateToken_WrongAlgorithm(t *testing.T) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
	}

	tokenObj := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := tokenObj.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	service := NewService("test-secret", 24*time.Hour)
	parsedClaims, err := service.ValidateToken(tokenString)
	assert.Nil(t, parsedClaims)
	assert.Error(t, err)
}
