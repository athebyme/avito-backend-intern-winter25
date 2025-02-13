package main

import (
	"avito-backend-intern-winter25/config"
	"avito-backend-intern-winter25/internal/handlers"
	"avito-backend-intern-winter25/internal/middleware"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"log"
)

const (
	configLocation = "config/config.yaml"
	migrationDir   = "file:./migrations"
)

func main() {
	cfg := &config.PostgresConfig{}

	rawCfg, err := cfg.Init(configLocation)
	if err != nil {
		log.Fatalf("error initilizating config : %v", err)
	}

	cfg, ok := rawCfg.(*config.PostgresConfig)
	if !ok {
		log.Fatalf("unexpected configuration file.")
	}

	connectionString := cfg.GetConnectionString()
	fmt.Println("Connection String:", connectionString)

	m, err := migrate.New(migrationDir, connectionString)
	if err != nil {
		log.Fatalf("Ошибка инициализации миграций: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Ошибка применения миграций: %v", err)
	}
	log.Println("Миграции успешно применены!")

	// Инициализация логгера
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	r := gin.New()

	// Мидлвари
	r.Use(
		middleware.Logging(logger),
		middleware.Prometheus(),
		gin.Recovery(),
	)

	// Эндпоинт для метрик Prometheus

	r := gin.Default()

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	authGroup := r.Group("/api")
	authGroup.Use(middleware.Auth())
	{
		authGroup.GET("/info", handlers.UserHandler)
		authGroup.POST("/sendCoin", handlers.SendCoins)
		authGroup.GET("/buy/:item", handlers.BuyItem)
	}

	r.POST("/api/auth", handlers.Auth)

	r.Run(":8080")
}
