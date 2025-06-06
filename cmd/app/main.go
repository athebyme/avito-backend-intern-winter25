package main

import (
	"avito-backend-intern-winter25/config"
	"avito-backend-intern-winter25/internal/handlers"
	"avito-backend-intern-winter25/internal/middleware"
	"avito-backend-intern-winter25/internal/services"
	"avito-backend-intern-winter25/internal/services/jwt"
	"avito-backend-intern-winter25/internal/storage/postgres"
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"os"
	"runtime"
	"time"
)

const (
	configLocation = "config/config.yaml"
	migrationDir   = "file://migrations"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	cfg, err := config.LoadConfig(configLocation)
	if err != nil {
		logger.Error("Error loading config", zap.Error(err))
	}

	jwtService := jwt.NewService(cfg.JWT.SecretKey, cfg.JWT.TokenLifetime)
	connectionString := cfg.Postgres.GetConnectionString()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Db,
	})

	ctx := context.Background()
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Connected to Redis")

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		logger.DPanic("Error connecting to database", zap.Error(err))
		os.Exit(1)
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	usrRepo := postgres.NewUserRepository(db)
	purchaseRepo := postgres.NewPurchaseRepository(db)
	merchRepo := postgres.NewMerchRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)

	usrService := services.NewUserService(usrRepo, jwtService, redisClient)
	merchService := services.NewMerchService(merchRepo, purchaseRepo, usrRepo, db)
	transactionService := services.NewTransactionService(db, usrRepo, transactionRepo)

	handler := handlers.NewHandler(usrService, merchService, transactionService, *logger)

	r := gin.Default()
	r.Use(
		middleware.Logging(logger),
		middleware.Prometheus(),
		gin.Recovery(),
	)

	handler.SetupRoutes(r, jwtService)

	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		logger.Fatal("Failed to run server", zap.Error(err))
	}
}
