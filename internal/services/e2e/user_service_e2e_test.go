package e2e

import (
	"avito-backend-intern-winter25/config"
	"avito-backend-intern-winter25/internal/handlers"
	"avito-backend-intern-winter25/internal/services"
	"avito-backend-intern-winter25/internal/services/jwt"
	"avito-backend-intern-winter25/internal/storage/postgres"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoginE2E(t *testing.T) {
	ctx := context.Background()
	cfg := setupTestConfig()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Db,
	})

	db, err := sql.Open("postgres", cfg.Postgres.GetConnectionString())
	require.NoError(t, err, "Failed to connect to database")

	userRepo := postgres.NewUserRepository(db)
	jwtService := jwt.NewService(cfg.JWT.SecretKey, cfg.JWT.TokenLifetime)
	userService := services.NewUserService(userRepo, jwtService, redisClient)

	purchaseRepo := postgres.NewPurchaseRepository(db)
	merchRepo := postgres.NewMerchRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	merchService := services.NewMerchService(merchRepo, purchaseRepo, userRepo, db)
	transactionService := services.NewTransactionService(db, userRepo, transactionRepo)

	handler := handlers.NewHandler(userService, merchService, transactionService)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.SetupRoutes(router, jwtService)

	t.Run("Register new user", func(t *testing.T) {
		username := fmt.Sprintf("testuser_%d", time.Now().UnixNano())
		password := "password123"

		reqBody := map[string]string{
			"username": username,
			"password": password,
		}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/auth", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var response map[string]string
		err = json.Unmarshal(resp.Body.Bytes(), &response)
		require.NoError(t, err)

		token, exists := response["token"]
		assert.True(t, exists, "Response should contain token")
		assert.NotEmpty(t, token, "Token should not be empty")

		user, err := userService.GetUserByUsername(ctx, username)
		require.NoError(t, err)
		assert.Equal(t, username, user.Username)
		assert.Equal(t, 1000, user.Coins, "New user should have 1000 coins")
	})

	t.Run("Login existing user", func(t *testing.T) {
		username := fmt.Sprintf("existing_user_%d", time.Now().UnixNano())
		password := "password123"

		reqBody := map[string]string{
			"username": username,
			"password": password,
		}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/auth", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, "/api/auth", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var response map[string]string
		err = json.Unmarshal(resp.Body.Bytes(), &response)
		require.NoError(t, err)

		token, exists := response["token"]
		assert.True(t, exists, "Response should contain token")
		assert.NotEmpty(t, token, "Token should not be empty")
	})

	t.Run("Login with invalid password", func(t *testing.T) {
		username := fmt.Sprintf("wrong_pass_user_%d", time.Now().UnixNano())
		password := "correct_password"

		reqBody := map[string]string{
			"username": username,
			"password": password,
		}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/auth", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)

		reqBody["password"] = "wrong_password"
		jsonBody, err = json.Marshal(reqBody)
		require.NoError(t, err)

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, "/api/auth", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)

		var errResponse map[string]string
		err = json.Unmarshal(resp.Body.Bytes(), &errResponse)
		require.NoError(t, err)

		errMsg, exists := errResponse["errors"]
		assert.True(t, exists, "Response should contain errors field")
		assert.Equal(t, "invalid password", errMsg)
	})

	t.Run("Verify user cache", func(t *testing.T) {
		username := fmt.Sprintf("cache_user_%d", time.Now().UnixNano())
		password := "password123"

		reqBody := map[string]string{
			"username": username,
			"password": password,
		}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/auth", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)

		cachedData, err := redisClient.Get(ctx, "user:"+username).Result()
		assert.NoError(t, err)
		assert.NotEmpty(t, cachedData, "User should be cached in Redis")

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, "/api/auth", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
	})

	if err := redisClient.Close(); err != nil {
		t.Logf("Failed to close Redis connection: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Logf("Failed to close database connection: %v", err)
	}
}

func setupTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
		JWT: config.JWTConfig{
			SecretKey:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			TokenLifetime: 24 * time.Hour,
		},
		Postgres: config.PostgresConfig{
			Host:     "db",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			DBName:   "shop",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Address:  "redis-test:6379",
			Password: "redis",
			Db:       0,
		},
	}
}
